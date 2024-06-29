import { createApp } from "vue/dist/vue.esm-browser.prod.js";
import { parseCookies } from "./common.js";
import { store } from "./store.js";

import ViewRecord from "./components/ViewRecord.ts";
import ViewRequest from "./components/ViewRequest.js";
import NewRecord from "./components/NewRecord.js";
import DomainLink from "./components/DomainLink.js";
import Experiments from "./components/Experiments.js";

const app = createApp({
  data() {
    return {
      domain: undefined,
      store: store,
    };
  },

  async mounted() {
    const cookies = parseCookies();
    const username = cookies["username"];
    if (username) {
      this.domain = username;
      await this.postLogin();
    }
    // add 'mounted' class to app element
    document.getElementById("app").classList.add("mounted");
  },

  computed: {
    records: function () {
      return this.store.records;
    },
    requests: function () {
      return this.store.requests;
    },
  },

  methods: {
    logout: function () {
      // clear cookies
      document.cookie = "username=; expires=Thu, 01 Jan 1970 00:00:01 GMT;";
      store.logout();
      this.domain = undefined;
    },

    clearRecords: async function () {
      if (confirm("Are you sure you want to delete all records?")) {
        await Promise.all(this.records.map(store.deleteRecord));
      }
    },

    clearRequests: async function () {
      if (confirm("Are you sure you want to delete all requests?")) {
        await store.deleteRequests(this.domain);
      }
    },

    postLogin: async function () {
      await store.init(this.domain);
    },
  },
});

app.component("record", ViewRecord);
app.component("view-request", ViewRequest);
app.component("new-record", NewRecord);
app.component("domain-link", DomainLink);
app.component("experiments", Experiments);

app.mount("#app");
