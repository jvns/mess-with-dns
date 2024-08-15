import { reactive } from "vue/dist/vue.esm-browser.prod.js";
import { parseCookies } from "./common.js";

interface Store {
    records: Record[]
    requests: Request[]
    ws: WebSocket
}

interface Record {
    id: string
    subdomain: string
    type: string
    ttl: string
    value: object // e.g. '{A: 1.2.3.4}' but the keys depend on the type
}

export const store: Store = reactive({
    records: [],
    requests: [],
    ws: undefined,

    async init() {
        await Promise.all([refreshRecords(), refreshRequests()]);
        openWebsocket();
    },
    async logout() {
        this.records = [];
        this.requests = [];
        this.ws.close();
    },
    async deleteRecord(record_id) {
        const url = '/records/' + record_id;
        const response = await fetch(url, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
            },
        });
        if (!response.ok) {
            alert('Error deleting record');
        }
        refreshRecords();
    },
    async createRecord(record) {
        // Make a copy before cleaning it up
        const response = await fetch('/records/', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(record),
        });
        refreshRecords();
        return response;
    },
    async updateRecord(id, record) {
        const url = '/records/' + id;
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(record),
        });
        refreshRecords();
        return response; 
    },

    async deleteRequests() {
        const response = await fetch('/requests', {
            method: 'DELETE',
        });
        refreshRequests();
        return response;
    },
});

async function refreshRecords() {
    const response = await fetch('/records/');
    store.records = await response.json();
}

async function refreshRequests() {
    const response = await fetch('/requests/');
    store.requests = await response.json();
}

function openWebsocket() {
    // use insecure socket on localhost
    var ws;
    const cookies = parseCookies();
    const username = cookies["username"];
    console.log("Opening websocket for", username);
    if (window.location.hostname === "localhost") {
        ws = new WebSocket("ws://localhost:8080/requeststream/" + username);
    } else {
        ws = new WebSocket("wss://" + window.location.host + "/requeststream/" + username);
    }
    store.ws = ws;
    ws.onmessage = (event) => {
        // ignore ping message
        if (event.data === "ping") {
            console.log("ping, ignoring");
            return;
        }
        const data = JSON.parse(event.data);
        store.requests.unshift(data);
    };
    ws.onclose = (e) => {
        console.log(
            "Websocket is closed. Reconnect will be attempted in 1 second.",
            e.reason,
        );
        setTimeout(() => {
            openWebsocket();
        }, 1000);
    };

    ws.onerror = (err) => {
        console.error(
            "Socket encountered error: ",
            err.message,
            "Closing socket",
        );
        ws.close();
    };
}

