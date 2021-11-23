import {
    convertRecord,
    fullName
} from '../common';

import * as schemas from '../schemas.json';
import template from './ViewRecord.html';

export default {
    template: template,
    props: ['record', 'domain'],
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
        fullName: function() {
            return fullName(this.record)
        },
        content: function() {
            let content = "";
            for (const key in this.record) {
                if (key == 'domain' || key == 'subdomain' || key == 'type' || key == 'ttl' || key == 'id') {
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
            await this.$parent.deleteRecord(this.record);
        },
        updateRecord: async function(data) {
            // TODO: terrible inconsistent naming, maybe fix, ugh
            data.subname = this.record.name;
            data.name = this.domain;
            const url = '/record/' + this.record.id;
            const record = convertRecord(data);
            const response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(record),
            });
            if (response.ok) {
                this.clicked = false;
                // update record in list
                this.$parent.updateRecord(this.record, this.updated_record);
                this.$parent.updateHash();
            } else {
                alert('Error updating record');
            }
        },
    },
};
