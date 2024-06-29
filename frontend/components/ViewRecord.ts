import { displayName } from '../common';
import { store } from '../store';

import template from './ViewRecord.html';

export default {
    template: template,
    props: ['record', 'domain'],
    data: function() {
        return {
            clicked: false,
        };
    },
    methods: {
        toggle: function() {
            this.clicked = !this.clicked;
        },
        displayName: function() {
            return displayName(this.record)
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
        confirmDelete: async function() {
            if (confirm('Are you sure you want to delete this record?')) {
                await store.deleteRecord(this.record);
            }
        },
        cancel: function() {
            this.clicked = false;
        },
    }};
