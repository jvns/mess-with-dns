package records

import (
	"context"
	"encoding/base64"
	"fmt"
	powerdns "github.com/joeig/go-powerdns/v3"
	"github.com/jvns/mess-with-dns/parsing"
	"net/http"
	"strings"
	"time"
)

type RecordService struct {
	pdns *powerdns.Client
}

func Init(url string, api_key string) RecordService {
	pdns := powerdns.NewClient(url, "localhost", map[string]string{"X-API-Key": api_key}, nil)
	return RecordService{pdns: pdns}
}

type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string {
	return e.Message
}

func newHTTPError(code int, err error) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: err.Error(),
	}
}

func (rs RecordService) DeleteRecord(ctx context.Context, username string, id string) *HTTPError {
	zone, err := rs.getOrCreateZone(ctx, username)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, err)
	}
	pdnsID, err := ParseID(id)
	if err != nil {
		return newHTTPError(http.StatusBadRequest, err)
	}
	err = zoneDelete(zone, pdnsID.Name, pdnsID.Type, pdnsID.Content)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, err)
	}
	err = rs.updateZone(ctx, username, zone)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, err)
	}
	return nil
}

func (rs RecordService) CreateRecord(ctx context.Context, username string, record map[string]string) *HTTPError {
	zone, err := rs.getOrCreateZone(ctx, username)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, err)
	}
	newRRset, err := parsing.ParseRecordRequest(record, username)
	if err != nil {
		return newHTTPError(http.StatusBadRequest, err)
	}
	fmt.Printf("%s %s %d %s\n", *newRRset.Name, *newRRset.Type, *newRRset.TTL, *newRRset.Records[0].Content)
	zoneAdd(zone, newRRset)
	err = rs.updateZone(ctx, username, zone)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, TranslateError(newRRset, err))
	}
	return nil
}

func (rs RecordService) UpdateRecord(ctx context.Context, username string, id string, record map[string]string) *HTTPError {
	zone, err := rs.getOrCreateZone(ctx, username)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, err)
	}
	pdnsID, err := ParseID(id)
	if err != nil {
		return newHTTPError(http.StatusBadRequest, err)
	}
	err = zoneDelete(zone, pdnsID.Name, pdnsID.Type, pdnsID.Content)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, err)
	}
	newRRset, err := parsing.ParseRecordRequest(record, username)
	if err != nil {
		return newHTTPError(http.StatusBadRequest, err)
	}
	zoneAdd(zone, newRRset)
	err = rs.updateZone(ctx, username, zone)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, TranslateError(newRRset, err))
	}
	return nil
}

func (rs RecordService) DeleteAllRecords(ctx context.Context, username string) *HTTPError {
	name := zoneName(username)
	err := rs.pdns.Zones.Delete(ctx, name)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, err)
	}
	return nil
}

func (rs RecordService) CreateZone(ctx context.Context, username string) (*powerdns.Zone, error) {
	zoneName := zoneName(username)
	kind := powerdns.NativeZoneKind
	zone := powerdns.Zone{
		Name:   &zoneName,
		Kind:   &kind,
		RRsets: []powerdns.RRset{},
	}
	_, err := rs.pdns.Zones.Add(ctx, &zone)
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

func (rs RecordService) getOrCreateZone(ctx context.Context, username string) (*powerdns.Zone, error) {
	zoneName := zoneName(username)
	zone, err := rs.pdns.Zones.Get(ctx, zoneName)
	if err != nil {
		return rs.CreateZone(ctx, username)
	}
	return zone, nil
}

func (rs RecordService) updateZone(ctx context.Context, username string, zone *powerdns.Zone) error {
	zoneName := zoneName(username)
	rrsets := []powerdns.RRset{}
	for _, rrset := range zone.RRsets {
		// only if changetype is replace
		if rrset.ChangeType != nil {
			rrsets = append(rrsets, rrset)
		}
	}
	err := rs.pdns.Records.Patch(ctx, zoneName, &powerdns.RRsets{rrsets})
	if err != nil {
		return err
	}
	return nil
}

func findRRsetIndex(zone *powerdns.Zone, name string, typ powerdns.RRType) (int, error) {
	for i, rrset := range zone.RRsets {
		if *rrset.Name == name && *rrset.Type == typ {
			return i, nil
		}
	}
	return 0, fmt.Errorf("Record not found: %s/%s", name, typ)
}

func zoneAdd(zone *powerdns.Zone, rrset *powerdns.RRset) {
	idx, err := findRRsetIndex(zone, *rrset.Name, *rrset.Type)
	replace := powerdns.ChangeTypeReplace
	if err != nil {
		rrset.ChangeType = &replace
		zone.RRsets = append(zone.RRsets, *rrset)
	} else {
		existing := &zone.RRsets[idx]
		existing.TTL = rrset.TTL
		existing.ChangeType = &replace
		existing.Records = append(existing.Records, rrset.Records...)
	}
}

func debugRRsets(zone *powerdns.Zone, num int) {
	fmt.Printf("rrsets %d: %d\n", num, len(zone.RRsets))
	for _, rrset := range zone.RRsets {
		fmt.Printf("-   %s %s [", *rrset.Name, *rrset.Type)
		for _, record := range rrset.Records {
			fmt.Printf("\"%s\" ", *record.Content)
		}
		fmt.Printf("]\n")
	}
}

func zoneDelete(zone *powerdns.Zone, name string, typ powerdns.RRType, content string) error {
	idx, err := findRRsetIndex(zone, name, typ)
	if err != nil {
		return fmt.Errorf("Record not found")
	}
	rrset := &zone.RRsets[idx]
	replace := powerdns.ChangeTypeReplace
	found := false
	for i := 0; i < len(rrset.Records); i++ {
		if *rrset.Records[i].Content == content {
			rrset.ChangeType = &replace
			rrset.Records = append(rrset.Records[:i], rrset.Records[i+1:]...)
			found = true
		}
	}
	if !found {
		return fmt.Errorf("Record not found")
	}
	return nil
}

var TLD string = "messwithdns.com."

func zoneName(username string) string {
	return fmt.Sprintf("%s.%s", username, TLD)
}

type Record struct {
	ID     string                 `json:"id"`
	Record parsing.RecordResponse `json:"record"`
}

func (rs RecordService) DeleteOldRecords(ctx context.Context, now time.Time) error {
	days := 7
	// Deletes any zones that haven't been updated in the last $DAYS days
	zones, err := rs.pdns.Zones.List(ctx)
	if err != nil {
		return fmt.Errorf("could not list zones: %v", err)
	}
	for _, zone := range zones {
		if *zone.Name == "messwithdns.com." {
			continue
		}
		date, err := ParseSerial(*zone.Serial)
		if err != nil {
			return fmt.Errorf("could not parse serial for %s: %v", *zone.Name, err)
		}
		if now.Sub(date).Hours() < float64(days*24) {
			continue
		}
		err = rs.pdns.Zones.Delete(ctx, *zone.Name)
		if err != nil {
			return fmt.Errorf("could not delete zone %s: %v", *zone.Name, err)
		}
	}
	return nil
}

func ParseSerial(serial uint32) (time.Time, error) {
	// serial format: YYYYMMDDNN
	serialStr := fmt.Sprintf("%010d", serial)
	date, err := time.Parse("20060102", serialStr[:8])
	if err != nil {
		return time.Time{}, err
	}
	return date, nil
}

func (rs RecordService) GetRecords(ctx context.Context, username string) ([]Record, *HTTPError) {
	zone, err := rs.getOrCreateZone(ctx, username)
	if err != nil {
		// not right, should probably be a 404
		return nil, newHTTPError(http.StatusInternalServerError, err)
	}
	// convert zone to RecordRequest
	records := []Record{}
	for _, rrset := range zone.RRsets {
		// filter out SOA and NS records
		//if *rrset.Type == powerdns.RRTypeSOA || *rrset.Type == powerdns.RRTypeNS {
		//	continue
		//}
		responses, err := parsing.RRsetToRecordResponse(&rrset)
		if err != nil {
			return nil, newHTTPError(http.StatusInternalServerError, err)
		}
		for _, resp := range responses {
			pdnsID := PdnsID{
				Name:    resp.DomainName,
				Type:    *rrset.Type,
				Content: resp.Content,
			}
			records = append(records, Record{
				ID:     pdnsID.String(),
				Record: resp,
			})
		}
	}
	return records, nil
}

type PdnsID struct {
	Name    string
	Content string
	Type    powerdns.RRType
}

func (id PdnsID) String() string {
	content := base64.StdEncoding.EncodeToString([]byte(id.Content))
	return fmt.Sprintf("%s|%s|%s", id.Name, id.Type, content)
}

func ParseID(id string) (*PdnsID, error) {
	// format of id: www.example.com|A|base64(content)
	parts := strings.SplitN(id, "|", 3)
	content, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}
	return &PdnsID{
		Name:    parts[0],
		Type:    powerdns.RRType(parts[1]),
		Content: string(content),
	}, nil
}
