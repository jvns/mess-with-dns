import * as schemas from "../schemas.json";
import template from "./NewRecord.html";
import { createRecord, updateRecord } from "../common.js";

export default {
  template: template,
  props: ["domain", "record"],
  data: function () {
    return {
      schemas: schemas,
      form_data: {},
      form_error: undefined,
    };
  },
  async mounted() {
    this.setFormState();
    // iterate over all inputs & selects and call setFormState on change
    const form = this.$refs.form;
    form.querySelectorAll("input, select").forEach((input) => {
      input.addEventListener("input", this.setFormState);
    });

    // wait a tick for form to render
    await new Promise((resolve) => setTimeout(resolve, 0));

    // set form's to record values if needed
    if (this.record) {
      const form = this.$refs.form;
      for (const key in this.record) {
        const input = form.elements[key];
        if (input) {
          input.value = this.record[key];
        } else {
          console.warn(`No input for key ${key}`);
        }
      }
    }
  },

  created: function () {
    // clear error when type is changed
    this.$watch("type", function () {
      this.error = undefined;
    });
  },

  methods: {
    setFormState: function () {
      const form = this.$refs.form;
      const formData = new FormData(form);
      this.form_data = Object.fromEntries(formData.entries());
    },

    cancel() {
      this.$parent.cancel();
    },

    getOptions() {
      if (this.record) {
        return [this.record.type];
      }
      return Object.keys(schemas).filter((key) => key != "default");
    },

    createOrUpdateRecord: async function (event) {
      this.setFormState();
      if (this.record) {
        await this.updateRecord(event);
      } else {
        await this.createRecord(event);
      }
    },

    updateRecord: async function () {
      const data = { ...this.form_data };
      const response = await updateRecord(this.record, data);
      if (response.ok) {
        this.$parent.$parent.refreshRecords();
        this.cancel();
      } else {
        this.form_error = await response.text();
      }
    },

    createRecord: async function (event) {
      const form = event.target;
      const data = { ...this.form_data };
      data.subdomain = data.subdomain.trim();
      data.domain = this.domain;
      const response = await createRecord(data);
      window.response = response;
      // check for errors
      if (response.status != 200) {
        this.form_error = await response.text();
        return;
      }
      this.$parent.refreshRecords();
      // clear form but keep type
      this.form_error = undefined;
      form.reset();
      this.setFormState();
      form.blur();
    },
  },
};
