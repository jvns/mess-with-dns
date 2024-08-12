$ORIGIN messwithdns.com.
@	       3600	IN SOA	mess-with-dns1.wizardzines.com. julia.wizardzines.com. 119088 3600 3600 7300 3600
@          3600 IN A    213.188.214.254
orange     3600 IN A    213.188.218.160
purple     3600 IN A    213.188.209.192
www        3600 IN A    213.188.214.254
_psl       3600 IN TXT  "https://github.com/publicsuffix/list/pull/1490"
test       3600 IN A    1.2.3.4
