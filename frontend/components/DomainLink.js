export default {
    template: `
<code>
  <a
    v-if="domain && domain.length > 0"
    target="_blank"
    :href="'http://' + subdomain + '.' + domain + '.messwithdns.com'"
    ><span class="text-green-700">{{subdomain}}.{{ domain }}</span
    >.messwithdns.com</a
  >
  <span v-else>
    <span class="text-green-700">your-domain</span>.messwithdns.com
  </span>
</code>
    `,
    props: ['domain', 'subdomain'],
}
