package polytools

import (
	"os"

	"github.com/Jeffail/gabs/v2"
	"github.com/bytedance/sonic"
)

//Polyfuncs was just designed to re-use certain functions from Gabs and OS but takes in pointers to allow memory efficiency
// Ownership + Borrowing is used here

func ParseJSON(sample *[]byte) (gabs.Container, error) {
	var gab interface{}
	if err := sonic.Unmarshal(*sample, &gab); err != nil {
		return *gabs.Wrap(&gab), err
	}
	return *gabs.Wrap(gab), nil
}

func ParseJSONFile(path string) (*gabs.Container, error) {

	cBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	container, err := ParseJSON(&cBytes)
	if err != nil {
		return nil, err
	}

	return &container, nil
}

// This is from the OS function. It does the same thing but data now takes in a pointer to make it use less memory
func WriteFile(name string, data *[]byte, perm os.FileMode) error {
	f, err := os.OpenFile(name, 1|64|512, perm)
	if err != nil {
		return err
	}
	_, err = f.Write(*data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}
