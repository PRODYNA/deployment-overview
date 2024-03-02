package publish

import "os"

func WriteFile(file string, data []byte) error {
	// Open file for writing
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	// Write data to file
	_, err = f.Write(data)
	if err != nil {
		return err
	}

	// Close file
	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}
