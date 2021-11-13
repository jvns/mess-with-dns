import Vue from 'vue/dist/vue.esm.js';
import VueFormulate from '@braid/vue-formulate'
import * as words from './words.json';
import * as schemas from './schemas.json';
import * as rrTypes from './rrTypes.json';

Vue.use(VueFormulate)

// reverse rrTypes
const rrTypesReverse = {};
for (const key in rrTypes) {
    rrTypesReverse[rrTypes[key]] = key;
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
        if (key == 'Target' || key == 'Mx' || key == 'Ns') {
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
            const predicates = words.predicates;
            const objects = words.objects;
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
