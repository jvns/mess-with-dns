import * as schemas from '../schemas.json';
import template from './NewRecord.html';
import {
    convertRecord
} from '../common.js';

export default {
    template: template,
    props: ['domain'],
    data: function() {
        return {
            type: 'A',
            data: undefined,
            schemas: schemas,
            // don't include 'default' in keys
            options: Object.keys(schemas).filter(key => key != 'default'),
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
            data.domain = this.domain;
            const record = convertRecord(data);
            console.log(record);
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
            this.$parent.updateHash();
            // clear form but keep type
            const type = this.data.type;
            document.activeElement.blur();
            this.$formulate.reset('new-record')
            this.data = {
                type: type
            };
        },
    }
}
