/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package metadata

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/palletone/go-palletone/common/log"
)

//var log = flogging.MustGetLogger("metadata")

// fileValidators are used as handlers to validate specific metadata directories
type fileValidator func(fileName string, fileBytes []byte) error

// Currently, the only metadata expected and allowed is for META-INF/statedb/couchdb/indexes.
var fileValidators = map[string]fileValidator{
	"META-INF/statedb/couchdb/indexes": couchdbIndexFileValidator,
}

// UnhandledDirectoryError is returned for metadata files in unhandled directories
type UnhandledDirectoryError struct {
	err string
}

func (e *UnhandledDirectoryError) Error() string {
	return e.err
}

// BadExtensionError is returned for metadata files with extension other than .json
type BadExtensionError struct {
	err string
}

func (e *BadExtensionError) Error() string {
	return e.err
}

// InvalidIndexContentError is returned for metadata files with invalid content
type InvalidIndexContentError struct {
	err string
}

func (e *InvalidIndexContentError) Error() string {
	return e.err
}

// ValidateMetadataFile checks that metadata files are valid
// according to the validation rules of the metadata directory (metadataType)
func ValidateMetadataFile(fileName string, fileBytes []byte, metadataType string) error {
	// Get the validator handler for the metadata directory
	fileValidator, ok := fileValidators[metadataType]

	// If there is no validator handler for metadata directory, return UnhandledDirectoryError
	if !ok {
		return &UnhandledDirectoryError{fmt.Sprintf("Metadata not supported in directory: %s", metadataType)}
	}

	// If the file is not valid for the given metadata directory, return the corresponding error
	err := fileValidator(fileName, fileBytes)
	if err != nil {
		return err
	}

	// file is valid, return nil error
	return nil
}

// couchdbIndexFileValidator implements fileValidator
func couchdbIndexFileValidator(fileName string, fileBytes []byte) error {

	ext := filepath.Ext(fileName)

	// if the file does not have a .json extension, then return as error
	if ext != ".json" {
		return &BadExtensionError{fmt.Sprintf("Index metadata file [%s] does not have a .json extension", fileName)}
	}

	// if the content does not validate as JSON, return err to invalidate the file
	boolIsJSON, indexDefinition := isJSON(fileBytes)
	if !boolIsJSON {
		return &InvalidIndexContentError{fmt.Sprintf("Index metadata file [%s] is not a valid JSON", fileName)}
	}

	// validate the index definition
	err := validateIndexJSON(indexDefinition)
	if err != nil {
		return &InvalidIndexContentError{fmt.Sprintf("Index metadata file [%s] is not a valid index definition: %s", fileName, err)}
	}

	return nil

}

// isJSON tests a string to determine if it can be parsed as valid JSON
func isJSON(s []byte) (bool, map[string]interface{}) {
	var js map[string]interface{}
	if err := json.Unmarshal(s, &js); err != nil {
		return false, nil
	}
	return true, js
}

func validateIndexJSON(indexDefinition map[string]interface{}) error {

	//flag to track if the "index" key is included
	indexIncluded := false

	//iterate through the JSON index definition
	for jsonKey, jsonValue := range indexDefinition {

		//create a case for the top level entries
		switch jsonKey {

		case "index":

			if reflect.TypeOf(jsonValue).Kind() != reflect.Map {
				return fmt.Errorf("Invalid entry, \"index\" must be a JSON")
			}

			err := processIndexMap(jsonValue.(map[string]interface{}))
			if err != nil {
				return err
			}

			indexIncluded = true

		case "ddoc":

			//Verify the design doc is a string
			if reflect.TypeOf(jsonValue).Kind() != reflect.String {
				return fmt.Errorf("Invalid entry, \"ddoc\" must be a string")
			}

			log.Debugf("Found index object: \"%s\":\"%s\"", jsonKey, jsonValue)

		case "name":

			//Verify the name is a string
			if reflect.TypeOf(jsonValue).Kind() != reflect.String {
				return fmt.Errorf("Invalid entry, \"name\" must be a string")
			}

			log.Debugf("Found index object: \"%s\":\"%s\"", jsonKey, jsonValue)

		case "type":

			if jsonValue != "json" {
				return fmt.Errorf("Index type must be json")
			}

			log.Debugf("Found index object: \"%s\":\"%s\"", jsonKey, jsonValue)

		default:

			return fmt.Errorf("Invalid Entry.  Entry %s", jsonKey)

		}

	}

	if !indexIncluded {
		return fmt.Errorf("Index definition must include a \"fields\" definition")
	}

	return nil

}

//processIndexMap processes an interface map and wraps field names or traverses
//the next level of the json query
func processIndexMap(jsonFragment map[string]interface{}) error {

	//iterate the item in the map
	for jsonKey, jsonValue := range jsonFragment {

		switch jsonKey {

		case "fields":

			switch jsonValueType := jsonValue.(type) {

			case []interface{}:

				//iterate the index field objects
				for _, itemValue := range jsonValueType {

					switch reflect.TypeOf(itemValue).Kind() {

					case reflect.String:
						//String is a valid field descriptor  ex: "color", "size"
						log.Debugf("Found index field name: \"%s\"", itemValue)

					case reflect.Map:
						//Handle the case where a sort is included  ex: {"size":"asc"}, {"color":"desc"}
						err := validateFieldMap(itemValue.(map[string]interface{}))
						if err != nil {
							return err
						}

					}
				}

			default:
				return fmt.Errorf("Expecting a JSON array of fields")
			}

		case "partial_filter_selector":

			//TODO - add support for partial filter selector, for now return nil
			//Take no other action, will be considered valid for now

		default:

			//if anything other than "fields" or "partial_filter_selector" was found,
			//return an error
			return fmt.Errorf("Invalid Entry.  Entry %s", jsonKey)

		}

	}

	return nil

}

//validateFieldMap validates the list of field objects
func validateFieldMap(jsonFragment map[string]interface{}) error {

	//iterate the fields to validate the sort criteria
	for jsonKey, jsonValue := range jsonFragment {

		switch jsonValue.(type) {

		case string:
			//Ensure the sort is either "asc" or "desc"
			if !(strings.ToLower(jsonValue.(string)) == "asc" || strings.ToLower(jsonValue.(string)) == "desc") {
				return fmt.Errorf("Sort must be either \"asc\" or \"desc\".  \"%s\" was found.", jsonValue)
			}
			log.Debugf("Found index field name: \"%s\":\"%s\"", jsonKey, jsonValue)

		default:
			return fmt.Errorf("Invalid field definition, fields must be in the form \"fieldname\":\"sort\"")

		}
	}

	return nil

}
