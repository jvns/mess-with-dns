export default {
    template: `
<code>
  <a
    v-if="domain && domain.length > 0"
    target="_blank"
    :href="'http://' + subdomain + '.' + domain + '.messwithdns.net'"
    >{{subdomain}}.{{ domain }}.messwithdns.net</a
  >
  <span v-else>
    <span class="text-green-700">your-domain</span>.messwithdns.net
  </span>
</code>
    `,
    props: ['domain', 'subdomain'],
}
