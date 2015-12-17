# from http://www.revsys.com/writings/quicktips/nat.html
ech0 1 > /proc/sys/net/ipv4/ip_forward

iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
iptables -A FORWARD -i eth0 -o tun66 -m state --state RELATED, ESTABLISHED -j ACCEPT
iptables -A FORWARD -i tun66 -o eth0 -j ACCEPT

