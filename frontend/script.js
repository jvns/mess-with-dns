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

    async mounted() {
        const cookies = parseCookies();
        const username = cookies['username'];
        if (username) {
            this.domain = username;
            await this.postLogin();
        }
        // add 'mounted' class to app element
        document.getElementById('app').classList.add('mounted');
    },

    methods: {
        logout: function() {
            // clear cookies
            document.cookie = 'username=; expires=Thu, 01 Jan 1970 00:00:01 GMT;';
            this.domain = undefined;
            this.requests = [];
            this.records = undefined;
            this.ws.close();
            this.websocketOpen = false;
        },
        clearRecords: async function() {
            if (confirm('Are you sure you want to delete all records?')) {
                await Promise.all(this.records.map(deleteRecord));
                await this.refreshRecords();
            }
        },
        clearRequests: async function() {
            if (confirm('Are you sure you want to delete all requests?')) {
                await deleteRequests(this.domain);
                await this.refreshRequests();
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

        openWebsocket: function() {
            // use insecure socket on localhost
            var ws;
            if (window.location.hostname === 'localhost') {
                ws = new WebSocket('ws://localhost:8080/requeststream');
            } else {
                ws = new WebSocket('wss://' + window.location.host + '/requeststream');
            }
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
            // refresh records and requests
            await Promise.all([
                this.refreshRecords(),
                this.refreshRequests()
            ]);
            // subscribe to stream for future requests
            this.openWebsocket();
        },
    }
});
