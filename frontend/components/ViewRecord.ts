import { displayName } from '../common';
import { store } from '../store';

import template from './ViewRecord.html';

export default {
    template: template,
    props: ['record', 'domain', 'id'],
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
            return displayName(this.domain, this.record)
        },
        content: function() {
            let content = "";
            for (const val of this.record.values) {
                content += val.value + " ";
            }
            return content;
        },
        confirmDelete: async function() {
            if (confirm('Are you sure you want to delete this record?')) {
                await store.deleteRecord(this.id);
            }
        },
        cancel: function() {
            this.clicked = false;
        },
    }};
