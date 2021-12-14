import rrTypes from './rrTypes.json';
import schemas from './schemas.json';

export interface Record {
    id: string
    domain: string
    subdomain: string
    type: string
    ttl: string
}

interface RecordUpdate {
    type: string
    ttl: string
}

export interface Header {
    Name: string
    Rrtype: number
    Class: number
    Ttl: number
    Rdlength: number
}

export interface GoRecord {
    Hdr: Header
}

export const rrTypesReverse = {};
for (const key in rrTypes) {
    rrTypesReverse[rrTypes[key]] = key;
}

export function displayName(record: Record) {
    if (record.subdomain == '@') {
        return record.domain + ".messwithdns.com";
    } else {
        return record.subdomain + '.' + record.domain + ".messwithdns.com";
    }
}

export function fullName(record: Record) {
    if (record.subdomain == '@') {
        return record.domain + ".messwithdns.com.";
    } else {
        return record.subdomain + '.' + record.domain + ".messwithdns.com.";
    }
}

function convertRecord(record: Record): GoRecord {
    // convert to api format
    // { "type": "A", "name": "example", "A": "93.184.216.34" }
    // =>
    // { "Hdr": { "Name": "example.messwithdns.com.", "Rrtype": 1, "Class": 1, "Ttl": 5, "Rdlength": 0 }, "A": "93.184.216.34" }
    const domainName = fullName(record);
    const newRecord: GoRecord = {
        Hdr: {
            Name: domainName,
            Rrtype: rrTypes[record.type],
            Class: 1,
            Ttl: parseInt(record.ttl),
            Rdlength: 0,
        },
    }
    // copy rest of fields from form directly
    for (const key in record) {
        if (key == 'Target' || key == 'Mx' || key == 'Ns' || key == 'Ptr' || key == 'Mbox') {
            // make sure it's a FQDN
            if (!record[key].endsWith('.')) {
                record[key] += '.';
            }
        }
        if (key == 'Txt') {
            // make sure it's a list of strings
            if (!Array.isArray(record[key])) {
                record[key] = [record[key]];
            }
        }
        // trim if it's a string
        if (typeof record[key] == 'string') {
            record[key] = record[key].trim();
        }
        if (key != 'domain' && key != 'subdomain' && key != 'type' && key != 'ttl') {
            // check if the type is 'number' in the schema
            const field = getSchemaField(record.type, key);
            if (field && field.type == 'number') {
                newRecord[key] = parseInt(record[key]);
            } else {
                newRecord[key] = record[key];
            }
        }
    }
    return newRecord;
}

function parseName(name: string): [string, string] {
    const parts = name.split('.');
    if (parts.length == 5) {
        return [parts[0], parts[1]]
    } else {
        return ['@', parts[0]]
    }
}

function transformRecord(id: string, record: GoRecord): Record {
    // { "Hdr": { "Name": "example.messwithdns.com.", "Rrtype": 1, "Class": 1, "Ttl": 5, "Rdlength": 0 }, "A": "
    // =>
    // { ttl: 5, name: "example", type: 'A' }
    const [subdomain, domain] = parseName(record.Hdr.Name)
    const basic = {
        id: id,
        ttl: record.Hdr.Ttl + '',
        domain: domain,
        subdomain: subdomain,
        type: rrTypesReverse[record.Hdr.Rrtype],
    };
    // copy rest of fields from record directly
    for (const key in record) {
        if (key != 'Hdr') {
            basic[key] = record[key];
        }
    }
    return basic;
}
export async function deleteRecord(record: Record) {
    const url = '/record/' + record.id;
    const response = await fetch(url, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
        },
    });
    if (!response.ok) {
        alert('Error deleting record');
    }
}

export async function updateRecord(old: Record, newRecord: RecordUpdate) {
    const url = '/record/' + old.id;
    // this merge thing seems a bit hacky
    const merged = {...old, ...newRecord};
    const goRecord = convertRecord(merged);
    const response = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(goRecord),
    });
    return response; 
}

export async function createRecord(record: Record) {
    const goRecord = convertRecord(record);
    const response = await fetch('/record/new', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(goRecord),
    });
    return response;
}

export async function getRecords(domain) {
    const response = await fetch('/domains/' + domain);
    const json = await response.json();
    // id is key, value is record
    const records = [];
    for (const r of json) {
        console.log(r);
        const record = transformRecord(r.id, r.record); 
        // parse key as int
        records.push(record);
    }
    return records;
}

export async function getRequests() {
    const response = await fetch('/requests');
    const json = await response.json();
    for (const event of json) {
        fixRequest(event);
    }
    return json;
}

export async function deleteRequests() {
    const response = await fetch('/requests', {
        method: 'DELETE',
    });
    return response;
}

export async function fixRequest(event) {
    event.request = JSON.parse(event.request);
    event.response = JSON.parse(event.response);
}

export function parseCookies() {
    const cookie = document.cookie;
    const cookies = cookie.split(';');
    const cookieObj = {};
    for (const c of cookies) {
        const cookie = c.split('=');
        cookieObj[cookie[0].trim()] = cookie[1];
    }
    return cookieObj;
}

function getSchemaField(type, key) {
    const fields = schemas[type];
    for (const field of fields) {
        if (field.name == key) {
            return field;
        }
    }
}
