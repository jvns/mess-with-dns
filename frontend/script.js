Vue.use(VueFormulate)

const schemas = {
  "A": [
    {
      "label": "IPv4 Address",
      "name": "A",
      "type": "text",
      "validation": "matches:/[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+/"
    }
  ],
  "AAAA": [
    {
      "label": "IPv6 Address",
      "name": "AAAA",
      "type": "text",
      "validation": "matches:/[0-9a-fA-F:]+/"
    }
  ],
  "CAA": [
    {
      "label": "Flag",
      "name": "Flag",
      "type": "number",
      "validation": "number|between:0,255"
    },
    {
      "label": "Tag",
      "name": "Tag",
      "type": "text",
      "validation": "required"
    },
    {
      "label": "CA domain name",
      "name": "Value",
      "type": "text",
      "validation": "required"
    }
  ],
  "CERT": [
    {
      "label": "Cert type",
      "name": "Type",
      "type": "number",
      "validation": "number|between:0,65535"
    },
    {
      "label": "Key tag",
      "name": "KeyTag",
      "type": "number",
      "validation": "number|between:0,65535"
    },
    {
      "label": "Algorithm",
      "name": "Algorithm",
      "type": "number",
      "validation": "number|between:0,255"
    },
    {
      "label": "Certificate",
      "name": "Certificate",
      "type": "text",
      "validation": "matches:/^[a-zA-Z0-9\\+\\/\\=]+$/"
    }
  ],
  "CNAME": [
    {
      "label": "Target",
      "name": "Target",
      "type": "text",
      "validation": "matches:/^([a-zA-Z0-9-]+\\.)*[a-zA-Z0-9]+\\.[a-zA-Z]+$/"
    }
  ],
  "DS": [
    {
      "label": "Key tag",
      "name": "KeyTag",
      "type": "number",
      "validation": "number|between:0,65535"
    },
    {
      "label": "Algorithm",
      "name": "Algorithm",
      "type": "number",
      "validation": "number|between:0,255"
    },
    {
      "label": "Digest type",
      "name": "DigestType",
      "type": "number",
      "validation": "number|between:0,255"
    },
    {
      "label": "Digest",
      "name": "Digest",
      "type": "text",
      "validation": "matches:/^[a-fA-F0-9]+$/"
    }
  ],
  "MX": [
    {
      "label": "Preference",
      "name": "Preference",
      "type": "number",
      "validation": "number|between:0,65535"
    },
    {
      "label": "Mail Server",
      "name": "Mx",
      "type": "text",
      "validation": "matches:/^([a-zA-Z0-9-]+\\.)*[a-zA-Z0-9]+\\.[a-zA-Z]+$/"
    }
  ],
  "NS": [
    {
      "label": "Nameserver ",
      "name": "Ns",
      "type": "text",
      "validation": "matches:/^([a-zA-Z0-9-]+\\.)*[a-zA-Z0-9]+\\.[a-zA-Z]+$/"
    }
  ],
  "PTR": [
    {
      "label": "Pointer",
      "name": "Ptr",
      "type": "text",
      "validation": "matches:/^([a-zA-Z0-9-]+\\.)*[a-zA-Z0-9]+\\.[a-zA-Z]+$/"
    }
  ],
  "SOA": [
    {
      "label": "Name Server",
      "name": "Ns",
      "type": "text",
      "validation": "matches:/^([a-zA-Z0-9-]+\\.)*[a-zA-Z0-9]+\\.[a-zA-Z]+$/"
    },
    {
      "label": "Email address",
      "name": "Mbox",
      "type": "text",
      "validation": "matches:/^([a-zA-Z0-9-]+\\.)*[a-zA-Z0-9]+\\.[a-zA-Z]+$/"
    },
    {
      "label": "Serial",
      "name": "Serial",
      "type": "number",
      "validation": "number"
    },
    {
      "label": "Refresh",
      "name": "Refresh",
      "type": "number",
      "validation": "number"
    },
    {
      "label": "Retry",
      "name": "Retry",
      "type": "number",
      "validation": "number"
    },
    {
      "label": "Expire",
      "name": "Expire",
      "type": "number",
      "validation": "number"
    },
    {
      "label": "Minimum TTL",
      "name": "Minttl",
      "type": "number",
      "validation": "number"
    }
  ],
  "SRV": [
    {
      "label": "Priority",
      "name": "Priority",
      "type": "number",
      "validation": "number|between:0,65535"
    },
    {
      "label": "Weight",
      "name": "Weight",
      "type": "number",
      "validation": "number|between:0,65535"
    },
    {
      "label": "Port",
      "name": "Port",
      "type": "number",
      "validation": "number|between:0,65535"
    },
    {
      "label": "Target",
      "name": "Target",
      "type": "text",
      "validation": "matches:/^([a-zA-Z0-9-]+\\.)*[a-zA-Z0-9]+\\.[a-zA-Z]+$/"
    }
  ],
  "TXT": [
    {
      "label": "Content",
      "name": "Txt",
      "type": "textarea",
      "validation": "required"
    }
  ],
  "URI": [
    {
      "label": "Priority",
      "name": "Priority",
      "type": "number",
      "validation": "number|between:0,65535"
    },
    {
      "label": "Weight",
      "name": "Weight",
      "type": "number",
      "validation": "number|between:0,65535"
    },
    {
      "label": "Target",
      "name": "Target",
      "type": "text",
      "validation": "required"
    }
  ]
};
const rrTypes = {
  "A": 1,
  "AAAA": 28,
  "AFSDB": 18,
  "ANY": 255,
  "APL": 42,
  "ATMA": 34,
  "AVC": 258,
  "AXFR": 252,
  "CAA": 257,
  "CDNSKEY": 60,
  "CDS": 59,
  "CERT": 37,
  "CNAME": 5,
  "CSYNC": 62,
  "DHCID": 49,
  "DLV": 32769,
  "DNAME": 39,
  "DNSKEY": 48,
  "DS": 43,
  "EID": 31,
  "EUI48": 108,
  "EUI64": 109,
  "GID": 102,
  "GPOS": 27,
  "HINFO": 13,
  "HIP": 55,
  "HTTPS": 65,
  "ISDN": 20,
  "IXFR": 251,
  "KEY": 25,
  "KX": 36,
  "L32": 105,
  "L64": 106,
  "LOC": 29,
  "LP": 107,
  "MAILA": 254,
  "MAILB": 253,
  "MB": 7,
  "MD": 3,
  "MF": 4,
  "MG": 8,
  "MINFO": 14,
  "MR": 9,
  "MX": 15,
  "NAPTR": 35,
  "NID": 104,
  "NIMLOC": 32,
  "NINFO": 56,
  "NS": 2,
  "NSAP-PTR": 23,
  "NSEC": 47,
  "NSEC3": 50,
  "NSEC3PARAM": 51,
  "NULL": 10,
  "NXT": 30,
  "None": 0,
  "OPENPGPKEY": 61,
  "OPT": 41,
  "PTR": 12,
  "PX": 26,
  "RKEY": 57,
  "RP": 17,
  "RRSIG": 46,
  "RT": 21,
  "Reserved": 65535,
  "SIG": 24,
  "SMIMEA": 53,
  "SOA": 6,
  "SPF": 99,
  "SRV": 33,
  "SSHFP": 44,
  "SVCB": 64,
  "TA": 32768,
  "TALINK": 58,
  "TKEY": 249,
  "TLSA": 52,
  "TSIG": 250,
  "TXT": 16,
  "UID": 101,
  "UINFO": 100,
  "UNSPEC": 103,
  "URI": 256,
  "X25": 19,
  "ZONEMD": 63
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
            options: Object.keys(schemas),
            error: undefined,
        };
    },
    created: function() {
        // clear error when type is changed
        this.$watch('type', function() {
            this.error = undefined;
        });
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
                this.error = await response.text();
                return;
            }
            this.error = undefined;
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
        if (key == 'Target') {
            // make sure it's a FQDN
            if (!record[key].endsWith('.')) {
                record[key] += '.';
            }
        }
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
        schemas: schemas,
        domain: undefined,
        records: undefined,
        words: undefined,
    },
    created: function() {
        // load words.json
        fetch('/words.json')
            .then(response => response.json())
            .then(json => {
                app.words = json;
            });
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
        setDomain: function(data) {
            this.domain = data.domain;
            window.location.hash = app.domain;
            updateHash()
        },
        randomSubdomain: function() {
            // predicate - object
            // return random word from words.json
            const predicates = this.words.predicates;
            const objects = this.words.objects;
            const predicate = predicates[Math.floor(Math.random() * predicates.length)];
            const object = objects[Math.floor(Math.random() * objects.length)];
            const domain = predicate + '-' + object;
            return domain;
        },
        goToRandom: function() {
            const domain = this.randomSubdomain();
            this.setDomain({domain: domain});
        },
    },
});

async function updateHash() {
    var hash = window.location.hash;
    if (hash.length == 0) {
        app.domain = undefined;
        return;
    }
    var domain = hash.substring(1);
    app.domain = domain;
    app.records = await app.getRecords(domain);
}

updateHash();
// update hash on change
window.onhashchange = updateHash;
