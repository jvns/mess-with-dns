import {
    updateRecord,
    deleteRecord,
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
                deleteRecord(this.record);
            }
            this.$parent.refreshRecords();
        },
        updateRecord: async function(data) {
            const response = await updateRecord(this.record, data);
            if (response.ok) {
                this.clicked = false;
                // update record in list
                this.$parent.refreshRecords();
            } else {
                alert('Error updating record');
            }
        },
    }};
