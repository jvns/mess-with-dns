import template from './Experiments.html';
export default {
    template: template,
    props: ['domain', 'subdomain'],
    mounted: function() {
        document.querySelectorAll('.experiment').forEach((exp) => {
            const summary = exp.querySelector('summary');
            const overlay = '<div class="details-modal-overlay"></div>'
            summary.insertAdjacentHTML('beforeend', overlay);
            const title = summary.innerText.split("\n")[0];
            const modal = exp.querySelector('.details-modal');
            const html = modal.innerHTML;
            modal.innerHTML = `
                <div class="details-modal-title"> ${title} </div>
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
