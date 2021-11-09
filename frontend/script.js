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

const schemas = {
    'A': [{
            'name': 'A',
            'label': 'IPv4 Address',
            'validation': "matches:/[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+\/",
            'validation-messages': {
                'matches': 'Invalid IPv4 Address'
            },
        },
    ],
    'AAAA': [{
            'name': 'AAAA',
            'label': 'IPv6 Address',
            'validation': 'required',
        },
    ],
    'CNAME': [
        {
            'name': 'target',
            'label': 'Target',
            'validation': 'required',
        },
    ],
    'MX': [
        {
            'name': 'priority',
            'label': 'Priority',
            'validation': 'required',
        },
        {
            'name': 'mail_server',
            'label': 'Mail server',
            'validation': 'required',
        },
    ],
    'NS': [
        {
            'name': 'nameserver',
            'label': 'Nameserver',
            'validation': 'required',
        },
    ],
    'TXT': [
        {
            'name': 'content',
            'label': 'Content',
            'type': 'textarea',
            'validation': 'required',
        },
    ],

};
// add name and ttl to every schema
for (const key in schemas) {
    schemas[key].push({
        'name': 'ttl',
        'label': 'TTL',
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
    },
});

Vue.component('new-record', {
    template: '#new-record',
    data: function() {
        return {
            schemas: schemas,
            type: 'A',
        };
    },

    methods: {
        createRecord: async function(data) {
            console.log(data);
            // { "type": "A", "name": "example", "A": "93.184.216.34" }
            // =>
            // { "Hdr": { "Name": "example.messwithdns.com.", "Rrtype": 1, "Class": 1, "Ttl": 5, "Rdlength": 0 }, "A": "93.184.216.34" }
            var record = {
                "Hdr": {
                    "Name": data.name + ".messwithdns.com.",
                    "Rrtype": rrTypes[data.type],
                    "Class": 1,
                    "Ttl": data.ttl,
                    "Rdlength": 0,
                },
            }
            // copy rest of fields from form directly
            for (var key in data) {
                if (key != 'name' && key != 'type' && key != 'ttl') {
                    record[key] = data[key];
                }
            }
            const response = await fetch('/record/new', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(record)
            });
            // check for errors
            if (response.status != 200) {
                console.log(response);
                return;
            }
            const json = await response.json();
            console.log(json);
        },
    }
});

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
