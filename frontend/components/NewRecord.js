import * as schemas from '../schemas.json';
import template from './NewRecord.html';
import {
    createRecord
} from '../common.js';

export default {
    template: template,
    props: ['domain'],
    data: function() {
        return {
            type: 'A',
            data: undefined,
            subdomain: undefined,
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
            // trim spaces
            data.subdomain = data.subdomain.trim();
            data.domain = this.domain;
            const response = await createRecord(data);
            window.response = response;
            // check for errors
            if (response.status != 200) {
                this.error = await response.text();
                return;
            }
            this.error = undefined;
            this.$parent.refreshRecords();
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
