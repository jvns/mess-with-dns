import Vue from 'vue/dist/vue.esm.js';

import VueFormulate from '@braid/vue-formulate'
Vue.use(VueFormulate)

import * as words from './words.json';
import * as schemas from './schemas.json';
import {
    transformRecord
} from './common.js';


import ViewRecord from './components/ViewRecord.ts';
import ViewRequest from './components/ViewRequest.js';
import NewRecord from './components/NewRecord.js';
import DomainLink from './components/DomainLink.js';
import Experiments from './components/Experiments.js';

Vue.component('record', ViewRecord);
Vue.component('view-request', ViewRequest);
Vue.component('new-record', NewRecord);
Vue.component('domain-link', DomainLink);
Vue.component('experiments', Experiments);


const vm = new Vue({
    el: '#app',
    data: {
        schemas: schemas,
        domain: undefined,
        events: [],
        records: undefined,
        sidebar: true,
    },

    methods: {
        getRecords: async function(domain) {
            const response = await fetch('/domains/' + domain);
            const json = await response.json();
            // id is key, value is record
            const records = [];
            for (var key in json) {
                const record = transformRecord(json[key]);
                // parse key as int
                record.id = parseInt(key);
                records.push(record);
            }
            return records;
        },

        clearAll: function() {
            if (confirm('Are you sure you want to delete all records?')) {
                for (var record of this.records) {
                    this.deleteRecord(record);
                }
            }
        },

        setDomain: function(data) {
            this.domain = data.domain;
            window.location.hash = this.domain;
            this.updateHash()
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
            this.setDomain({
                domain: domain
            });
        },
        deleteRecord: async function(record) {
            var url = '/record/' + record.id;
            var response = await fetch(url, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                },
            });
            if (response.ok) {
                // remove record from list
                var index = this.records.indexOf(record);
                this.records.splice(index, 1);
                return true;
            } else {
                alert('Error deleting record');
                return false;
            }
        },
        updateRecord: function(oldRecord, newRecord) {
            var index = this.records.indexOf(oldRecord);
            this.records[index] = newRecord;
        },

        updateHash: async function() {
            var hash = window.location.hash;
            if (hash.length == 0) {
                this.domain = undefined;
                return;
            }
            var domain = hash.substring(1);
            this.domain = domain;
            this.records = await this.getRecords(domain);
            this.events = [];
            // TODO: maybe initial events should be a different endpoint from ongoing
            // events
            const source = new EventSource('/events/' + domain);
            source.onmessage = event => {
                const data = JSON.parse(event.data);
                data.request = JSON.parse(data.request);
                data.response = JSON.parse(data.response);
                this.events.unshift(data);
            };
        },
    }
});
vm.updateHash();
window.onhashchange = vm.updateHash;
