import { reactive } from "vue/dist/vue.esm-browser.prod.js";
import punycode from 'punycode';

import { rrTypesReverse } from './common';
import rrTypes from './rrTypes.json';
import schemas from './schemas.json';

interface Store {
    domain: string
    records: Record[]
    requests: Request[]
    ws: WebSocket
}

interface Record {
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

interface Header {
    Name: string
    Rrtype: number
    Class: number
    Ttl: number
    Rdlength: number
}

interface GoRecord {
    Hdr: Header
    // I think there are actually more fields
    // here but it depends on the type
}

export const store: Store = reactive({
    domain: undefined,
    records: [],
    requests: [],
    ws: undefined,

    async init(domain) {
        this.domain = domain;
        await Promise.all([refreshRecords(), refreshRequests()]);
        openWebsocket();
    },
    async logout() {
        this.domain = undefined;
        this.records = [];
        this.requests = [];
        this.ws.close();
    },
    async deleteRecord(record) {
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
        refreshRecords();
    },
    async createRecord(record) {
        // Make a copy before cleaning it up
        record = {...record};
        record.domain = this.domain;
        record.subdomain = record.subdomain.trim();
        const goRecord = recordToGoRecord(record);
        const response = await fetch('/record/new', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(goRecord),
        });
        refreshRecords();
        return response;
    },
    async updateRecord(old, newRecord) {
        const url = '/record/' + old.id;
        // this merge thing seems a bit hacky
        const merged = {...old, ...newRecord};
        const goRecord = recordToGoRecord(merged);
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(goRecord),
        });
        refreshRecords();
        return response; 
    },

    async deleteRequests() {
        const response = await fetch('/requests', {
            method: 'DELETE',
        });
        refreshRequests();
        return response;
    },
});

async function getRecords() {
    const domain = this.domain;
    const response = await fetch('/domains/' + domain);
    const json = await response.json();
    // id is key, value is record
    const records = [];
    for (const r of json) {
        const record = recordFromGoRecord(r.id, r.record); 
        // parse key as int
        records.push(record);
    }
    return records;
}



async function getRequests() {
    const response = await fetch('/requests');
    const json = await response.json();
    for (const event of json) {
        fixRequest(event);
    }
    return json;
}



async function refreshRecords() {
    store.records = await getRecords();
}

async function refreshRequests() {
    store.requests = await getRequests();
}

async function fixRequest(event) {
    event.request = JSON.parse(event.request);
    event.response = JSON.parse(event.response);
}

function openWebsocket() {
    // use insecure socket on localhost
    var ws;
    if (window.location.hostname === "localhost") {
        ws = new WebSocket("ws://localhost:8080/requeststream");
    } else {
        ws = new WebSocket("wss://" + window.location.host + "/requeststream");
    }
    store.ws = ws;
    ws.onmessage = (event) => {
        // ignore ping message
        if (event.data === "ping") {
            console.log("ping, ignoring");
            return;
        }
        const data = JSON.parse(event.data);
        fixRequest(data);
        store.requests.unshift(data);
    };
    ws.onclose = (e) => {
        console.log(
            "Websocket is closed. Reconnect will be attempted in 1 second.",
            e.reason,
        );
        setTimeout(() => {
            openWebsocket();
        }, 1000);
    };

    ws.onerror = (err) => {
        console.error(
            "Socket encountered error: ",
            err.message,
            "Closing socket",
        );
        ws.close();
    };
}

function parseName(name: string): [string, string] {
    // split on '.' and trim last 3 segments
    const parts = name.split('.');
    if (parts.length == 4) {
        return ['@', parts[0]]
    } else {
        const subdomain = parts.slice(0, parts.length - 4).join('.');
        const domain = parts[parts.length-4];
        return [subdomain, domain];
    }
}

function recordFromGoRecord(id: string, record: GoRecord): Record {
    // { "Hdr": { "Name": "example.messwithdns.com.", "Rrtype": 1, "Class": 1, "Ttl": 5, "Rdlength": 0 }, "A": "
    // =>
    // { ttl: 5, name: "example", type: 'A' }
    const decoded = punycode.toUnicode(record.Hdr.Name);
    const [subdomain, domain] = parseName(decoded)
    const basic = {
        id: id,
        ttl: record.Hdr.Ttl + '',
        domain: domain,
        subdomain: subdomain,
        type: rrTypesReverse[record.Hdr.Rrtype],
    };
    // copy rest of fields from record directly
    for (const key in record) {
        if (key == 'Hdr') {
            continue;
        } else if (key == 'Txt') {
            // join array of 255 char strings
            basic[key] = record[key].join('');
        } else {
            basic[key] = record[key];
        }
    }
    return basic;
}

function fullName(record: Record) {
    if (record.subdomain == '@') {
        return record.domain + ".messwithdns.com.";
    } else {
        return record.subdomain + '.' + record.domain + ".messwithdns.com.";
    }
}

function recordToGoRecord(record: Record): GoRecord {
    // convert to api format
    // { "type": "A", "name": "example", "A": "93.184.216.34" }
    // =>
    // { "Hdr": { "Name": "example.messwithdns.com.", "Rrtype": 1, "Class": 1, "Ttl": 5, "Rdlength": 0 }, "A": "93.184.216.34" }
    const domainName = fullName(record);
    const punycoded = punycode.toASCII(domainName);
    const newRecord: GoRecord = {
        Hdr: {
            Name: punycoded,
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
            // split into array of 255 char strings
            const txt = record[key];
            const txtArray = [];
            for (let i = 0; i < txt.length; i += 255) {
                txtArray.push(txt.substring(i, i + 255));
            }
            record[key] = txtArray;
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

function getSchemaField(type, key) {
    const fields = schemas[type];
    for (const field of fields) {
        if (field.name == key) {
            return field;
        }
    }
}
