package engine

import (
	"archive/zip"
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/coinbase/dexter/engine/helpers"
	"github.com/coinbase/dexter/facts"
	"github.com/coinbase/dexter/tasks"

	log "github.com/Sirupsen/logrus"
	"github.com/fatih/color"
)

//
// An investigation is an instruction for some Dexter daemons
// to run some tasks.  The Task list defines the tasks and
// their argumetns, while the Scope defines facts that must
// be true about the host in order for the investigation to
// be in scope.
//
type Investigation struct {
	ID             string
	TaskList       map[string][]string
	Scope          map[string][]string
	KillContainers bool
	KillHost       bool
	Issuer         Signature
	Approvers      []Signature
	RecipientNames []string
}

//
// A signature consists of the name of the investigator who did
// the signing, and the signature data.
//
type Signature struct {
	Name string
	Data []byte
}

//
// A decryption payload contains all the information needed for an
// investigator to decrypt an investigation, a nonce for data
// encryption and the encrypted data encryption key.
//
type DecryptionPayload struct {
	Nonce                      []byte
	EncryptedDataEncryptionKey []byte
}

//
// Decrypt the encrypted data encryption key using the local investigator's
// key.  A password function is used to retrieve the password for the local
// investigator's private key.
//
func (payload DecryptionPayload) GetEncryptionKey(passwordFunc func() string) []byte {
	var data []byte
	var err error
	for {
		data, err = rsa.DecryptOAEP(sha256.New(), rand.Reader, helpers.LoadLocalKey(passwordFunc), payload.EncryptedDataEncryptionKey, []byte{})
		if err == nil {
			return data
		} else {
			color.HiRed("Decryption error in DecryptionPayload: " + err.Error())
			os.Exit(1)
		}
	}
	return data
}

//
// List the names of the investigators who approved an investigation.
//
func (investigation *Investigation) ApproverNames() []string {
	names := []string{}
	for _, sig := range investigation.Approvers {
		names = append(names, sig.Name)
	}
	return names
}

//
// Return the path on the local filesystem for the zipped report that
// resulted from this investigation.
//
func (investigation *Investigation) ReportZip() string {
	return os.TempDir() + "/DexterReport-" + investigation.ID + ".zip"
}

//
// Return the local filesystem path that is being used to write report
// artifacts during this investigation.
//
func (investigation *Investigation) ReportDirectory() string {
	return os.TempDir() + "/DexterReport-" + investigation.ID + "/"
}

func (investigation *Investigation) validate() error {
	// Verify the issuer has a valid signature
	if !investigation.validateSignature(investigation.Issuer) {
		return errors.New("issuer signature invalid")
	}

	// Load the tasks to run as defined in the TaskList
	if investigation.countValidTasks() == 0 {
		return errors.New("unable to load any tasks for investigation")
	}

	// Determine if the facts defined in the scope are relevent to this system
	for attribute, value := range investigation.Scope {
		factChecker, exists := facts.Get(attribute)
		if !exists {
			return errors.New("investigation attempts to check non-existent fact " + attribute)
		}
		if factChecker.Private {
			factChecker.Salt = investigation.ID
		}
		inScope := factChecker.Assert(value)
		if !inScope {
			return errors.New("host is not in scope, fact " + attribute + " does not apply")
		}
	}

	// Verify this action has been approved with +n consensus
	if !investigation.consensusRequirementsMet() {
		return errors.New("investigation has not yet reached consensus")
	}

	return nil
}

func (investigation *Investigation) run() {
	err := os.MkdirAll(filepath.FromSlash(investigation.ReportDirectory()), 0700)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "engine.run",
			"error": err.Error(),
			"path":  investigation.ReportDirectory(),
		}).Error("unable to create report directory")
		return
	}

	log.WithFields(log.Fields{
		"at":            "engine.run",
		"investigation": investigation.ID,
	}).Info("running investigation")
	for taskName, taskArgs := range investigation.TaskList {
		if task, ok := tasks.Tasks[taskName]; ok {
			task.Run(investigation.ReportDirectory(), taskArgs)
		} else {
			log.WithFields(log.Fields{
				"at":   "engine.taskListToTask",
				"name": taskName,
			}).Error("task name is not a known task")
		}
	}
	log.WithFields(log.Fields{
		"at":            "engine.run",
		"investigation": investigation.ID,
	}).Info(fmt.Sprintf("finished %d tasks", len(investigation.TaskList)))
}

func (investigation *Investigation) uniqueApprovers() []Signature {
	knownNames := make(map[string]bool)
	set := make([]Signature, 0)
	for _, sig := range investigation.Approvers {
		if sig.Name == investigation.Issuer.Name {
			log.WithFields(log.Fields{
				"at":   "engine.uniqueApprovers",
				"name": sig.Name,
			}).Error("issuer cannot also be approver")
			continue
		}
		if _, known := knownNames[sig.Name]; !known {
			set = append(set, sig)
			knownNames[sig.Name] = true
		}
	}
	return set
}

//
// Return the number of signatures on an investigation that are from a unique
// set of investigators and are valid.  This is equivalent to the current
// consensus level.
//
func (investigation *Investigation) ValidUniqueApprovers() int {
	signatures := investigation.uniqueApprovers()
	achieved := 0
	for _, sig := range signatures {
		if investigation.validateSignature(sig) {
			achieved += 1
		} else {
			log.WithFields(log.Fields{
				"at":            "engine.consensusRequirementsMet",
				"name":          sig.Name,
				"investigation": investigation.ID,
			}).Error("approver signature invalid")
		}
	}
	return achieved
}

func (investigation *Investigation) consensusRequirementsMet() bool {
	return investigation.ValidUniqueApprovers() >= investigation.MinimumConsensus()
}

//
// Each task has different consensus requirements, return the highest value
// from all the tasks.  That will be the amount of consensus required for
// this investigation.
//
func (investigation *Investigation) MinimumConsensus() int {
	required := 1
	for taskName, _ := range investigation.TaskList {
		task, ok := tasks.Tasks[taskName]
		if !ok {
			log.WithFields(log.Fields{
				"at":        "engine.MinimumConsensus",
				"task_name": taskName,
			}).Error("named task not found")
			continue
		}
		if task.ConsensusRequirement > required {
			required = task.ConsensusRequirement
		}
	}
	return required
}

func (investigation *Investigation) allSignaturesValid() bool {
	if !investigation.validateSignature(investigation.Issuer) {
		return false
	}
	for _, approver := range investigation.Approvers {
		if !investigation.validateSignature(approver) {
			return false
		}
	}
	return true
}

func (investigation *Investigation) validateSignature(sig Signature) bool {
	publicKey, err := GetPublicKeyForInvestigator(sig.Name)
	if err != nil {
		return false
	}
	err = rsa.VerifyPSS(publicKey, crypto.SHA256, investigation.digest(), sig.Data, &rsa.PSSOptions{})
	return err == nil
}

func (investigation *Investigation) digest() []byte {
	// Create a data blob of all the data sent in an investigation
	// that is not the signature data.  This allows all parties to
	// verify the signatures of the same investigation data.
	blob := make([]byte, 0)
	blob = append(blob, []byte(investigation.ID)...)
	// maps must always be in the same order so the hash is the same
	blob = append(blob, orderedMapData(investigation.TaskList)...)
	blob = append(blob, orderedMapData(investigation.Scope)...)
	if investigation.KillContainers {
		blob = append(blob, 0x01)
	} else {
		blob = append(blob, 0x00)
	}
	if investigation.KillHost {
		blob = append(blob, 0x01)
	} else {
		blob = append(blob, 0x00)
	}
	blob = append(blob, []byte(investigation.Issuer.Name)...)
	for _, recipient := range investigation.RecipientNames {
		blob = append(blob, []byte(recipient)...)
	}

	// Create a SHA-256 hash of this data and return it
	sum := sha256.Sum256(blob)
	return sum[:]
}

// Take all the map data and return it in a consistent order
// for use in the investigation digest
func orderedMapData(data map[string][]string) []byte {
	keys := []string{}
	for k, _ := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ret := []byte{}
	for _, str := range keys {
		ret = append(ret, []byte(str)...)
		args := data[str]
		for _, arg := range args {
			ret = append(ret, []byte(arg)...)
		}
	}
	return ret
}

func (investigation *Investigation) report() {
	log.WithFields(log.Fields{
		"at":            "engine.report",
		"investigation": investigation.ID,
	}).Info("reporting investigation")

	// create a zip of the entire report directory
	investigation.zip()
	// encrypt that zip file to each recipient
	for _, investigator := range investigation.RecipientNames {
		decryptionPayload := investigation.encrypt(investigator)
		log.WithFields(log.Fields{
			"at":            "engine.report",
			"investigation": investigation.ID,
			"recipient":     investigator,
		}).Info("uploading report")

		//s3 upload encrypted zip
		hostname, err := os.Hostname()
		if err != nil {
			log.WithFields(log.Fields{
				"at":            "engine.report",
				"error":         err.Error(),
				"investigation": investigation.ID,
			}).Error("unable to retrieve hostname")
			continue
		}
		encryptedZip, err := os.Open(filepath.FromSlash(investigation.ReportZip()) + ".enc")
		if err != nil {
			log.WithFields(log.Fields{
				"at":            "engine.report",
				"error":         err.Error(),
				"investigation": investigation.ID,
				"investigator":  investigator,
			}).Error("error opening encrypted report for upload")
			continue
		}
		reportUploadPath := "reports/" + investigation.ID + "-" + hostname + "." + investigator + ".zip.enc"
		err = helpers.UploadS3File(reportUploadPath, encryptedZip)
		if err != nil {
			log.WithFields(log.Fields{
				"at":            "engine.report",
				"error":         err.Error(),
				"investigation": investigation.ID,
				"user":          investigator,
			}).Error("unable to upload encrypted zip")
			continue
		}
		// s3 upload decryption package
		decryptionData, err := json.Marshal(decryptionPayload)
		if err != nil {
			log.WithFields(log.Fields{
				"at":            "engine.report",
				"error":         err.Error(),
				"investigation": investigation.ID,
				"user":          investigator,
			}).Error("unable to marshal decryption payload")
			continue
		}
		decryptionPayloadPath := "reports/" + investigation.ID + "-" + hostname + "." + investigator + ".decrypt"
		err = helpers.UploadS3File(decryptionPayloadPath, bytes.NewReader(decryptionData))
		if err != nil {
			log.WithFields(log.Fields{
				"at":            "engine.report",
				"error":         err.Error(),
				"investigation": investigation.ID,
				"user":          investigator,
			}).Error("unable to upload decryption payload")
			continue
		}
	}
}

//
// Encrypt an investigation for a specific investigator, returning the encrypted
// zip as well as the encrypted data encryption key
//
func (investigation Investigation) encrypt(user string) DecryptionPayload {
	clearZipData, err := ioutil.ReadFile(investigation.ReportZip())
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "engine.encrypt",
			"error": err.Error(),
			"file":  investigation.ReportZip(),
		}).Fatal("unable to open report zip for encryption")
	}
	key, nonce, ciphertext := aesgcmEncrypt(clearZipData)
	err = ioutil.WriteFile(investigation.ReportZip()+".enc", ciphertext, 0644)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "engine.encrypt",
			"error": err.Error(),
			"file":  investigation.ReportZip() + ".enc",
		}).Fatal("unable to open file for encrypted zip")
	}

	userPubKey, err := GetPublicKeyForInvestigator(user)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "Engine.encrypt",
			"error": err.Error(),
		}).Error("error getting public key for encrypted report")
		return DecryptionPayload{}
	}
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, userPubKey, key, []byte{})
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "engine.encrypt",
			"error": err.Error(),
		}).Fatal("encryption error")
	}

	return DecryptionPayload{
		Nonce: nonce,
		EncryptedDataEncryptionKey: encryptedKey,
	}
}

func aesgcmEncrypt(cleartext []byte) (key, nonce, ciphertext []byte) {
	key = make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		log.WithFields(log.Fields{
			"at":    "engine.aesgcmEncrypt",
			"error": err.Error(),
		}).Fatal("unable to generate random key")
	}
	nonce = make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.WithFields(log.Fields{
			"at":    "engine.aesgcmEncrypt",
			"error": err.Error(),
		}).Fatal("unable to generate random nonce")
	}

	// Create AES-GCM block cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "engine.aesgcmEncrypt",
			"error": err.Error(),
		}).Fatal("error calling aes.NewCipher")
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "engine.aesgcmEncrypt",
			"error": err.Error(),
		}).Fatal("error calling cipher.NewGCM")
	}
	return key, nonce, aesgcm.Seal(nil, nonce, cleartext, nil)
}

func (investigation *Investigation) zip() {
	out, err := os.Create(filepath.FromSlash(investigation.ReportZip()))
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "engine.zip",
			"error": err.Error(),
			"file":  investigation.ReportZip(),
		}).Error("unable to create report zip file")
		return
	}
	zipWriter := zip.NewWriter(out)
	err = filepath.Walk(investigation.ReportDirectory(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.WithFields(log.Fields{
				"at":    "engine.zip",
				"error": err.Error(),
				"path":  path,
			}).Error("error walking path")
			return err
		}
		if info.IsDir() {
			return nil
		}
		zipPath := path[len(investigation.ReportDirectory()):]
		file, err := zipWriter.Create(zipPath)
		if err != nil {
			log.WithFields(log.Fields{
				"at":    "engine.zip",
				"error": err.Error(),
				"path":  zipPath,
			}).Error("unable to create file in zip")
			return nil
		}
		src, err := os.Open(filepath.FromSlash(path))
		if err != nil {
			log.WithFields(log.Fields{
				"at":    "engine.zip",
				"error": err.Error(),
			}).Error("error opening report file for zip")
		}
		_, err = io.Copy(file, src)
		if err != nil {
			log.WithFields(log.Fields{
				"at":    "engine.zip",
				"error": err.Error(),
			}).Error("error writing report file into zip stream")
		}
		return nil
	})

	if err != nil {
		log.WithFields(log.Fields{
			"at":    "engine.zip",
			"error": err.Error(),
			"path":  investigation.ReportDirectory(),
		}).Error("error walking report path")
	}
	zipWriter.Close()
	out.Close()
}

func (investigation *Investigation) countValidTasks() int {
	count := 0
	for taskName, _ := range investigation.TaskList {
		if _, ok := tasks.Tasks[taskName]; ok {
			count += 1
		}
	}
	return count
}

func (investigation *Investigation) Sign(privateKey *rsa.PrivateKey) {
	hash := investigation.digest()
	sig, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, hash, &rsa.PSSOptions{})
	if err != nil {
		color.HiRed("Error signing investigation: " + err.Error())
		os.Exit(1)
	}
	investigation.Issuer.Data = sig
}

func (investigation *Investigation) Approve(privateKey *rsa.PrivateKey) {
	hash := investigation.digest()
	sig, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, hash, &rsa.PSSOptions{})
	if err != nil {
		color.HiRed("Error signing investigation: " + err.Error())
		os.Exit(1)
	}
	investigation.Approvers = append(investigation.Approvers, Signature{
		Name: LocalInvestigatorName(),
		Data: sig,
	})
}

//
// Get a slice of strings that are printable versions of the facts on this investigation.
//
func (investigation *Investigation) ScopeFactsStrings() []string {
	set := []string{}
	for fact, args := range investigation.Scope {
		checker, ok := facts.Get(fact)
		if !ok {
			color.HiRed("attempted to print fact that doesn't exist, should not be possible.  Different versions of Dexter?")
		} else {
			set = append(set, helpers.StringWithArgs(fact, args, checker.Private))
		}
	}
	return set
}

//
// Get a single string that represents all facts on this investigation.
//
func (investigation *Investigation) ScopeFactsToString() string {
	return strings.Join(investigation.ScopeFactsStrings(), ", ")
}

//
// Upload this investigation to S3.
//
func (investigation *Investigation) Upload() error {
	investigationBytes, _ := json.MarshalIndent(investigation, "", "  ")
	uploadPath := "investigations/" + investigation.ID + "." + LocalInvestigatorName()
	return helpers.UploadS3File(uploadPath, bytes.NewReader(investigationBytes))
}

//
// Lookup an investigation by ID, or partial ID.
//
func InvestigationByID(uuid string) (Investigation, error) {
	full, err := helpers.ResolveUUID(uuid)
	if err != nil {
		return Investigation{}, err
	}
	all := CurrentInvestigations()
	for _, inv := range all {
		if inv.ID == full {
			return inv, nil
		}
	}
	return Investigation{}, errors.New("UUID resolved but investigation not found")
}

//
// Lookup an investigation by ID, or partial ID, using an already downloaded list of investigation.
//
func InvestigationByIDWithCache(cache []Investigation, uuid string) (Investigation, error) {
	full, err := helpers.ResolveUUID(uuid)
	if err != nil {
		return Investigation{}, err
	}
	for _, inv := range cache {
		if inv.ID == full {
			return inv, nil
		}
	}
	return Investigation{}, errors.New("UUID resolved but investigation not found in cache")
}

//
// Download all investigations currently on S3.
//
func CurrentInvestigations() []Investigation {
	knownInvestigations := make(map[string]Investigation)

	investigations, err := helpers.ListS3Path("investigations/")
	if err != nil {
		color.HiRed(err.Error())
	}

	for _, filename := range investigations {
		data, err := helpers.GetS3File(filename)
		if err != nil {
			color.HiRed(err.Error())
		}
		var inv = Investigation{}
		err = json.Unmarshal(data, &inv)
		if err != nil {
			color.HiRed("unable to unmarshal investigation json: " + err.Error())
			continue
		}
		if !inv.allSignaturesValid() {
			color.HiRed("investigation contains invalid signatures")
			continue
		}
		if check, ok := knownInvestigations[inv.ID]; ok {
			if len(inv.Approvers) > len(check.Approvers) {
				knownInvestigations[inv.ID] = inv
			}
		} else {
			knownInvestigations[inv.ID] = inv
		}
	}
	set := []Investigation{}
	for _, v := range knownInvestigations {
		set = append(set, v)
	}
	return set
}
