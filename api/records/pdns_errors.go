package records

import (
	"fmt"
	powerdns "github.com/joeig/go-powerdns/v3"
	"strings"
)

func TranslateError(rrset *powerdns.RRset, err error) error {

	errorString := err.Error()
	name := *rrset.Name
	typ := *rrset.Type
	content := *rrset.Records[0].Content

	// RRset test.pear5.messwithdns.com. IN CNAME: Conflicts with pre-existing RRset
	if strings.Contains(errorString, "Conflicts with pre-existing RRset") {
		return fmt.Errorf("Error: can't create record for %s: CNAME records aren't allowed to coexist with other records", name)
	}
	// Duplicate record in RRset test.pear5.messwithdns.com. IN A with content "1.2.3.5"
	if strings.Contains(errorString, "Duplicate record in RRset") {
		return fmt.Errorf("Error: there's already a record with name %s, type %s, and content %s", name, typ, content)
	}
	// RRset test2.pear5.messwithdns.com. IN CNAME has more than one record
	if strings.Contains(errorString, "has more than one record") {
		return fmt.Errorf("Error: a name is only allowed to have one %s record, and %s already has one", typ, name)
	}
	return err
}
