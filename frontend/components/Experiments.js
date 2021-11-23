import template from './Experiments.html';
export default {
    template: template,
    props: ['domain', 'subdomain'],
    methods: {
        currDomain: function() {
            return this.domain || 'your-domain';
        },
    },
}
