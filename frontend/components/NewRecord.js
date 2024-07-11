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
      for (const field of this.record.values) {
        const input = form.elements['value_' + field.name];
        if (input) {
          input.value = field.value;
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
      this.form_data = {
          "values": [],
      }
      for (const [key, value] of formData.entries()) {
          // if it starts with 'value_` it's a value field
          if (key == 'ttl') {
              this.form_data[key] = parseInt(value);
          } else if (key.startsWith("value_")) {
              const real_key = key.slice(6);
              this.form_data["values"].push({
                  'name': real_key,
                  'value': value,
              });
          } else {
              this.form_data[key] = value
          }
      }
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
