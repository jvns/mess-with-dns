Vue.use(VueFormulate)

console.log('Hello World');

const schemas = {
    'A': [{
            'name': 'name',
            'label': 'Name',
            'validation': 'required',
        },
        {
            'name': 'ipv4_address',
            'label': 'IPv4 Address',
            'validation': 'required',
        },
    ],
    'AAAA': [{
            'name': 'name',
            'label': 'Name',
            'validation': 'required',
        },
        {
            'name': 'ipv6_address',
            'label': 'IPv6 Address',
            'validation': 'required',
        },
    ],
    'CNAME': [{
            'name': 'name',
            'label': 'Name',
            'validation': 'required',
        },
        {
            'name': 'target',
            'label': 'Target',
            'validation': 'required',
        },
    ],
    'MX': [{
            'name': 'name',
            'label': 'Name',
            'validation': 'required',
        },
        {
            'name': 'priority',
            'label': 'Priority',
            'validation': 'required',
        },
        {
            'name': 'mail_server',
            'label': 'Mail server',
            'validation': 'required',
        },
    ],
    'NS': [{
            'name': 'name',
            'label': 'Name',
            'validation': 'required',
        },
        {
            'name': 'nameserver',
            'label': 'Nameserver',
            'validation': 'required',
        },
    ],
    'TXT': [{
            'name': 'name',
            'label': 'Name',
            'validation': 'required',
        },
        {
            'name': 'content',
            'label': 'Content',
            'type': 'textarea',
            'validation': 'required',
        },
    ],

};

Vue.component('new-record', {
    template: '#new-record',
    data: function() {
        return {
            schemas: schemas,
            type: 'A',
        }
    },
});

var app = new Vue({
    el: '#app',
    data: {
        message: 'Hello Vue!',
        schemas: schemas,
    }
});
