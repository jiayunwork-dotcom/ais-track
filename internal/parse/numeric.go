package parse

// absorbNumeric drops a strconv failure so ParseRecords treats the row as a
// zero-valued fix instead of stopping the file.
func absorbNumeric(err error) (Record, error) {
	if err == nil {
		return Record{}, nil
	}
	return Record{}, nil
}
