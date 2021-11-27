import Vue from 'vue/dist/vue.esm.js';

import VueFormulate from '@braid/vue-formulate'
Vue.use(VueFormulate)

import * as words from './words.json';
import * as schemas from './schemas.json';
import {
    getRecords,
    getRequests,
    fixRequest,
    deleteRecord
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
        clearAll: function() {
            if (confirm('Are you sure you want to delete all records?')) {
                for (var record of this.records) {
                    deleteRecord(record);
                }
                this.refreshRecords();
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
        refreshRecords: async function() {
            this.records = await getRecords(this.domain);
        },

        updateHash: async function() {
            var hash = window.location.hash;
            if (hash.length == 0) {
                this.domain = undefined;
                return;
            }
            var domain = hash.substring(1);
            this.domain = domain;
            this.refreshRecords();
            // get past requests
            this.events = await getRequests(this.domain);
            // subscribe to stream for future requests
            const source = new EventSource('/requeststream/' + domain);
            source.onmessage = event => {
                const data = JSON.parse(event.data);
                fixRequest(data);
                this.events.unshift(data);
            };
        },
    }
});
vm.updateHash();
window.onhashchange = vm.updateHash;
