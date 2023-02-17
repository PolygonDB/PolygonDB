package polysync

import (
	"fmt"
	"sync"

	"github.com/Jeffail/gabs/v2"

	utils "github.com/JewishLewish/PolygonDB/GoPackage/utilities/polyFuncs"
)

var (
	Databases = &AtomicDatabase{
		data: make(map[string][]byte),
	}
)

type AtomicDatabase struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func (ad *AtomicDatabase) Load(location string) ([]byte, bool) {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	value, ok := ad.data[location]
	if !ok {
		return nil, false
	}

	return value, true
}

func (ad *AtomicDatabase) Store(location string, value []byte) {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	ad.data[location] = value
}

func Datacheck(location *string, database *gabs.Container) error {
	if value, ok := Databases.Load(*location); ok {
		*database, _ = utils.ParseJSON(&value)
		value = nil
	} else {
		var dataerr error
		*database, dataerr = data(location)
		if dataerr != nil {
			return dataerr
		}
	}
	return nil
}

func data(location *string) (gabs.Container, error) {

	value, err := utils.ParseJSONFile("databases/" + *location + "/database.json")
	if err != nil {
		go fmt.Println("Error unmarshalling Database JSON:", err)
	}
	Databases.Store(*location, value.Bytes())
	return *value, nil
}
