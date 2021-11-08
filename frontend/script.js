Vue.use(VueFormulate)

console.log('Hello World');

var app = new Vue({
    el: '#app',
    data: {
        message: 'Hello Vue!',
        type: 'A',
        schemas: {
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
        },
    }
});
