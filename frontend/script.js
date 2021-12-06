import Vue from 'vue/dist/vue.esm.js';

import VueFormulate from '@braid/vue-formulate'
Vue.use(VueFormulate)

import * as words from './words.json';
import * as schemas from './schemas.json';
import {
    getRecords,
    getRequests,
    deleteRequests,
    fixRequest,
    deleteRecord,
    parseCookies
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
        requests: [],
        records: undefined,
        ws: undefined,
        sidebar: true,
        websocketOpen: false,
    },

    mounted() {
        const cookies = parseCookies();
        const username = cookies['username'];
        if (username) {
            this.domain = username;
            this.postLogin();
        }
    },

    methods: {
        clearRecords: function() {
            if (confirm('Are you sure you want to delete all records?')) {
                for (var record of this.records) {
                    deleteRecord(record);
                }
                this.refreshRecords();
            }
        },
        clearRequests: function() {
            if (confirm('Are you sure you want to delete all requests?')) {
                deleteRequests(this.domain);
                this.refreshRequests();
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
            this.records = await getRecords();
        },
        refreshRequests: async function() {
            this.requests = await getRequests();
        },

        openWebsocket: async function() {
            const ws = new WebSocket('ws://' + window.location.host + '/requeststream');
            ws.addEventListener('open', () => {
                    this.websocketOpen = true;
            });
            ws.onmessage = (event) => {
                const data = JSON.parse(event.data);
                fixRequest(data);
                this.requests.unshift(data);
            };
            ws.onclose = (e) => {
                console.log('Websocket is closed. Reconnect will be attempted in 1 second.', e.reason);
                this.websocketOpen = false;
                setTimeout(() => {
                    this.openWebsocket();
                }, 1000);
            };

            ws.onerror = (err) => {
                console.error('Socket encountered error: ', err.message, 'Closing socket');
                ws.close();
            };
        },
        postLogin: async function() {
            this.refreshRecords();
            // get past requests
            this.refreshRequests();
            // subscribe to stream for future requests
            this.openWebsocket();
        },
    }
});
