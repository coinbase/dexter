package util

//
// Check if the string in the check variable is a member
// of the set of strings.
//
func StringsInclude(set []string, check string) bool {
	for _, member := range set {
		if member == check {
			return true
		}
	}
	return false
}

//
// Given a slice of strings, return the slice without the
// string in the rm variable, if it exists in the set.
//
func StringsSubtract(set []string, rm string) []string {
	newSet := []string{}
	for _, member := range set {
		if member != rm {
			newSet = append(newSet, member)
		}
	}
	return newSet
}

//
// Merge two string slices, only adding values from
// set2 that were not already present in set1.  The
// resulting slice will not be unique if there are
// repeat entries in set1, but repeat entries in set2
// will only be appended once.
//
func AppendUnique(set1, set2 []string) []string {
	finalSet := set1
	set1Map := make(map[string]bool)
	for _, member := range set1 {
		set1Map[member] = true
	}
	for _, member := range set2 {
		if _, ok := set1Map[member]; !ok {
			finalSet = append(finalSet, member)
		}
	}
	return finalSet
}
