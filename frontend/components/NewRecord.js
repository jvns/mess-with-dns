import * as schemas from "../schemas.json";
import template from "./NewRecord.html";
import { store } from "../store.js";

function getSchemaField(type, key) {
    const fields = schemas[type];
    for (const field of fields) {
        if (field.name == key) {
            return field;
        }
    }
}

export default {
  template: template,
  props: ["record", "domain", "id"],
  data: function () {
    return {
      schemas: schemas,
      form_data: {},
      form_error: undefined,
    };
  },
  async mounted() {
    this.setFormState(); // TODO: do we need this?
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
        } else if (key != "content" & key != "domain_name") {
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
      this.form_data = Object.fromEntries(formData.entries())
    },

    cancel() {
      this.$parent.cancel();
    },

    getOptions() {
      if (this.record) {
        return [this.record.type];
      }
      const types = Object.keys(schemas).sort();
      return types.filter((key) => key != "default");
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
      const response = await store.updateRecord(this.id, this.form_data);
      if (response.ok) {
        this.cancel();
      } else {
        this.form_error = await response.text();
      }
    },

    createRecord: async function (event) {
      const form = event.target;
      const response = await store.createRecord(this.form_data);
      if (response.status != 200) {
        this.form_error = await response.text();
        return;
      }
      // clear form
      this.form_error = undefined;
      form.reset();
      this.setFormState();
      form.blur();
    },
  },
};
