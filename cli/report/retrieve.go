package report

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/coinbase/dexter/cli/cliutil"
	"github.com/coinbase/dexter/engine"
	"github.com/coinbase/dexter/engine/helpers"

	log "github.com/sirupsen/logrus"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

//
// A ReportFile struct contains all the metadata contained in a report filename.
//
type ReportFile struct {
	ID        string
	Hostname  string
	Recipient string
}

func (file *ReportFile) getDecryptionPayload() engine.DecryptionPayload {
	decryptFile := "reports/" + file.ID + "-" + file.Hostname + "." + file.Recipient + ".decrypt"
	decryptData, err := helpers.GetS3File(decryptFile)
	if err != nil {
		color.HiRed("error getting file from S3: " + err.Error())
		os.Exit(1)
	}
	var decryptPayload engine.DecryptionPayload
	err = json.Unmarshal(decryptData, &decryptPayload)
	if err != nil {
		color.HiRed("error unmarshalling decryption payload: " + err.Error())
		os.Exit(1)
	}
	return decryptPayload
}

func (file *ReportFile) getEncryptedBlob() []byte {
	encryptedZipFile := "reports/" + file.ID + "-" + file.Hostname + "." + file.Recipient + ".zip.enc"
	data, err := helpers.GetS3File(encryptedZipFile)
	if err != nil {
		color.HiRed("error getting file from S3: " + err.Error())
		os.Exit(1)
	}
	return data
}

func retrieveReport(cmd *cobra.Command, args []string) {
	uuid, err := helpers.ResolveUUID(args[0])
	if err != nil {
		color.HiRed(err.Error())
		os.Exit(1)
	}
	name := engine.LocalInvestigatorName()

	files := filterFiles(uuid, name, ReportFiles())
	for _, file := range files {
		dataEncryptionKey := file.getDecryptionPayload().GetEncryptionKey(cliutil.CollectPassword)
		decryptedZip := decryptZip(file.getEncryptedBlob(), dataEncryptionKey, file.getDecryptionPayload().Nonce)
		reader, err := zip.NewReader(bytes.NewReader(decryptedZip), int64(len(decryptedZip)))
		if err != nil {
			color.HiRed("error creating zip reader for report: " + err.Error())
			os.Exit(1)
		}
		for _, zf := range reader.File {
			dir := "DexterReport-" + uuid + "/" + file.Hostname + "/" + path.Dir(zf.Name)
			err = os.MkdirAll(filepath.FromSlash(dir), 0700)
			if err != nil {
				color.HiRed("error creating report directory: " + err.Error())
				continue
			}
			file := dir + "/" + path.Base(zf.Name)
			src, err := zf.Open()
			if err != nil {
				color.HiRed("error opening zipped file: " + err.Error())
				continue
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(src)
			err = ioutil.WriteFile(file, buf.Bytes(), 0644)
			if err != nil {
				color.HiRed("error writing report file: " + err.Error())
				continue
			}
			src.Close()
		}
	}
}

func decryptZip(ciphertext []byte, key []byte, nonce []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		color.HiRed("decryption error creating cipher: " + err.Error())
		os.Exit(1)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		color.HiRed("decryption error creating GCM block: " + err.Error())
		os.Exit(1)
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		color.HiRed("decryption error: " + err.Error())
		os.Exit(1)
	}
	return plaintext
}

func filterFiles(uuid, user string, set []ReportFile) []ReportFile {
	filtered := make([]ReportFile, 0)
	for _, file := range set {
		if file.ID == uuid && file.Recipient == user {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

//
// List all files in the Dexter S3 bucket reports directory.
//
func ReportFiles() []ReportFile {
	files := make([]ReportFile, 0)
	stored := make(map[string]bool)

	filenames, err := helpers.ListS3Path("reports/")
	if err != nil {
		color.HiRed("unable to get report file from S3: " + err.Error())
		os.Exit(1)
	}

	re := regexp.MustCompile(`reports/(.+?)-(.+)\.(.+)\.zip\.enc`)
	for _, filename := range filenames {
		if !strings.HasSuffix(filename, ".zip.enc") {
			continue
		}
		matches := re.FindStringSubmatch(filename)
		if len(matches) < 4 {
			log.WithFields(log.Fields{
				"at":       "report.CurrentReports",
				"filename": filename,
			}).Error("regex mismatch on filename")
			continue
		}
		uuid := matches[1]
		hostname := matches[2]
		recipientName := matches[3]

		key := uuid + "." + hostname + "." + recipientName
		if _, present := stored[key]; !present {
			files = append(
				files,
				ReportFile{
					ID:        uuid,
					Hostname:  hostname,
					Recipient: recipientName,
				},
			)
			stored[key] = true
		}
	}
	return files
}
