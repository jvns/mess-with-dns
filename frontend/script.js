Vue.use(VueFormulate)

console.log('Hello World');

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
}

const schemas = {
    'A': [{
            'name': 'name',
            'label': 'Name',
            'validation': 'required',
        },
        {
            'name': 'A',
            'label': 'IPv4 Address',
            'validation': "matches:/[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+\/",
            'validation-messages': {
                'matches': 'Invalid IPv4 Address'
            },
        },
    ],
    'AAAA': [{
            'name': 'name',
            'label': 'Name',
            'validation': 'required',
        },
        {
            'name': 'AAAA',
            'label': 'IPv6 Address',
            'validation': 'required',
        },
    ],
    'CNAME': [{
            'name': 'CNAME',
            'label': 'Name',
            'validation': 'required',
        },
        {
            'name': 'target',
            'label': 'Target',
            'validation': 'required',
        },
    ],
    'MX': [{
            'name': 'name',
            'label': 'Name',
            'validation': 'required',
        },
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
    'NS': [{
            'name': 'name',
            'label': 'Name',
            'validation': 'required',
        },
        {
            'name': 'nameserver',
            'label': 'Nameserver',
            'validation': 'required',
        },
    ],
    'TXT': [{
            'name': 'name',
            'label': 'Name',
            'validation': 'required',
        },
        {
            'name': 'content',
            'label': 'Content',
            'type': 'textarea',
            'validation': 'required',
        },
    ],

};

Vue.component('new-record', {
    template: '#new-record',
    data: function() {
        return {
            schemas: schemas,
            type: 'A',
        }
    },

    methods: {
        createRecord: function(data) {
            console.log(data);
            // { "type": "A", "name": "help", "A": "1.2.3.4" }
            // =>
            // { "Hdr": { "Name": "example.messwithdns.com.", "Rrtype": 1, "Class": 1, "Ttl": 5, "Rdlength": 0 }, "A": "93.184.216.34" }
            var record = {
                "Hdr": {
                    "Name": data.name,
                    "Rrtype": rrTypes[data.type],
                    "Class": 1,
                    "Ttl": data.ttl,
                    "Rdlength": 0,
                },
            }
            // copy rest of fields from data
            for (var key in data) {
                if (key != 'name' && key != 'type' && key != 'ttl') {
                    record[key] = data[key];
                }
            }
            console.log(record);
        },
    }
});

var app = new Vue({
    el: '#app',
    data: {
        message: 'Hello Vue!',
        schemas: schemas,
    }
});
