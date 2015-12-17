
ip route add default via 192.168.3.1 dev tun66 table hof
ip rule add fwmark 1 table hof
# mark these packets so that iproute can route it through wlan-route
iptables -A OUTPUT -t mangle -o eth0 -p tcp --dport 80 -j MARK --set-mark 1
# now rewrite the src-addr
iptables -A POSTROUTING -t nat -o tun66 -p tcp --dport 80 -j SNAT --to 192.168.3.2 
