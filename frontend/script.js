Vue.use(VueFormulate)


const rrTypes = {
    'A': 1,
    'NS': 2,
    'MD': 3,
    'MF': 4,
    'CNAME': 5,
    'SOA': 6,
    'MB': 7,
    'MG': 8,
    'MR': 9,
    'NULL': 10,
    'WKS': 11,
    'PTR': 12,
    'HINFO': 13,
    'MINFO': 14,
    'MX': 15,
    'TXT': 16,
    'RP': 17,
    'AFSDB': 18,
    'X25': 19,
    'ISDN': 20,
    'RT': 21,
    'NSAP': 22,
    'NSAP-PTR': 23,
    'SIG': 24,
    'KEY': 25,
    'PX': 26,
    'GPOS': 27,
    'AAAA': 28,
    'LOC': 29,
    'NXT': 30,
    'EID': 31,
    'NIMLOC': 32,
    'SRV': 33,
    'ATMA': 34,
    'NAPTR': 35,
    'KX': 36,
    'CERT': 37,
    'A6': 38,
    'DNAME': 39,
    'SINK': 40,
    'OPT': 41,
    'APL': 42,
    'DS': 43,
    'SSHFP': 44,
    'IPSECKEY': 45,
    'RRSIG': 46,
    'NSEC': 47,
    'DNSKEY': 48,
    'DHCID': 49,
    'NSEC3': 50,
    'NSEC3PARAM': 51,
    'TLSA': 52,
    'SMIMEA': 53,
    'HIP': 55,
    'NINFO': 56,
    'RKEY': 57,
    'TALINK': 58,
    'CDS': 59,
    'CDNSKEY': 60,
    'OPENPGPKEY': 61,
    'CSYNC': 62,
    'SPF': 99,
    'UINFO': 100,
    'UID': 101,
    'GID': 102,
    'UNSPEC': 103,
    'NID': 104,
    'L32': 105,
    'L64': 106,
    'LP': 107,
    'EUI48': 108,
    'EUI64': 109,
    'TKEY': 249,
    'TSIG': 250,
    'IXFR': 251,
    'AXFR': 252,
    'MAILB': 253,
    'MAILA': 254,
    'ANY': 255,
    'URI': 256,
    'CAA': 257,
};

// reverse rrTypes
const rrTypesReverse = {};
for (const key in rrTypes) {
    rrTypesReverse[rrTypes[key]] = key;
}

// add name and ttl to every schema
for (const key in schemas) {
    schemas[key].push({
        'name': 'ttl',
        'label': 'TTL',
        'type': 'number',
        'validation': 'required',
    });
}

Vue.component('record', {
    template: '#view-record',
    props: ['record'],
    data: function() {
        return {
            schemas: schemas,
            clicked: false,
            updated_record: {},
        };
    },
    methods: {
        toggle: function() {
            if (!this.clicked) {
                this.updated_record = JSON.parse(JSON.stringify(this.record));
            }
            this.clicked = !this.clicked;
        },
        content: function() {
            var content = "";
            for (const key in this.record) {
                if (key == 'name' || key == 'type' || key == 'ttl' || key == 'id') {
                    continue;
                }
                content += this.record[key] + " ";
            }
            return content;
        },
        confirmDelete: function() {
            if (confirm('Are you sure you want to delete this record?')) {
                this.deleteRecord();
            }
        },
        deleteRecord: async function() {
            var url = '/record/' + this.record.id;
            var response = await fetch(url, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                },
            });
            if (response.ok) {
                // remove record from list
                var index = app.records.indexOf(this.record);
                app.records.splice(index, 1);
            } else {
                alert('Error deleting record');
            }
        },
        updateRecord: async function(data) {
            var url = '/record/' + this.record.id;
            const record = convertRecord(data);
            var response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(record),
            });
            if (response.ok) {
                // update record in list
                var index = app.records.indexOf(this.record);
                app.records[index] = this.updated_record;
                this.clicked = false;
                updateHash();
            } else {
                alert('Error updating record');
            }
        },
    },
});

Vue.component('new-record', {
    template: '#new-record',
    props: ['domain'],
    data: function() {
        return {
            schemas: schemas,
            type: 'A',
            data: undefined,
        };
    },

    methods: {
        createRecord: async function(data) {
            // { "type": "A", "name": "example", "A": "93.184.216.34" }
            // =>
            // { "Hdr": { "Name": "example.messwithdns.com.", "Rrtype": 1, "Class": 1, "Ttl": 5, "Rdlength": 0 }, "A": "93.184.216.34" }
            data.name = this.domain;
            const record = convertRecord(data);
            const response = await fetch('/record/new', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(record)
            });
            // check for errors
            if (response.status != 200) {
                // TODO: handle errors
                console.log(response);
                return;
            }
            updateHash();
            // clear form but keep type
            this.data = {type: this.data.type};
        },
    }
});

function convertRecord(record) {
    // convert to api format
    // { "type": "A", "name": "example", "A": "93.184.216.34" }
    // =>
    // { "Hdr": { "Name": "example.messwithdns.com.", "Rrtype": 1, "Class": 1, "Ttl": 5, "Rdlength": 0 }, "A": "93.184.216.34" }
    var newRecord = {
        "Hdr": {
            "Name": record.name + ".messwithdns.com.",
            "Rrtype": rrTypes[record.type],
            "Class": 1,
            "Ttl": parseInt(record.ttl),
            "Rdlength": 0,
        },
    }
    // copy rest of fields from form directly
    for (var key in record) {
        if (key != 'name' && key != 'type' && key != 'ttl') {
            // check if the type is 'number' in the schema
            const field = getSchemaField(record.type, key);
            if (field.type == 'number') {
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


var app = new Vue({
    el: '#app',
    data: {
        message: 'Hello Vue!',
        schemas: schemas,
        domain: undefined,
        records: undefined,
    },
    methods: {
        getRecords: async function(domain) {
            const response = await fetch('/domains/' + domain);
            const json = await response.json();
            // transform records
            // id is key, value is record
            const records = [];
            for (var key in json) {
                const record = this.transformRecord(json[key]);
                // parse key as int
                record.id = parseInt(key);
                records.push(record);
            }
            return records;
        },
        transformRecord: function(record) {
            // { "Hdr": { "Name": "example.messwithdns.com.", "Rrtype": 1, "Class": 1, "Ttl": 5, "Rdlength": 0 }, "A": "
            // =>
            // { ttl: 5, name: "example", type: 'A' }
            var basic = {
                ttl: record.Hdr.Ttl,
                name: record.Hdr.Name.split('.')[0],
                type: rrTypesReverse[record.Hdr.Rrtype],
            };
            // copy rest of fields from record directly
            for (var key in record) {
                if (key != 'Hdr') {
                    basic[key] = record[key];
                }
            }
            return basic;
        },
    },
});

async function updateHash() {
    var hash = window.location.hash;
    if (hash.length == 0) {
        return;
    }
    var domain = hash.substring(1);
    app.domain = domain;
    app.records = await app.getRecords(domain);
}

updateHash();
// update hash on change
window.onhashchange = updateHash;
