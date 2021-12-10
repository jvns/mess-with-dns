import template from './Experiments.html';
export default {
    template: template,
    props: ['domain', 'subdomain'],
    mounted: function() {
        document.querySelectorAll('.experiment').forEach((exp) => {
            const summary = exp.querySelector('summary');
            const overlay = '<div class="details-modal-overlay"></div>'
            summary.insertAdjacentHTML('beforeend', overlay);
            const h3 = exp.querySelector('summary h3');
            const modal = exp.querySelector('.details-modal');
            const html = modal.innerHTML;
            modal.innerHTML = `
                <div class="details-modal-title"> ${h3.innerHTML} </div>
                <div class="details-modal-close"> </div>
                <div class="details-modal-content">
                    ${html}
                </div>
            `;
        });
    },
    methods: {
        currDomain: function() {
            return this.domain || 'your-domain';
        },
    },
}
