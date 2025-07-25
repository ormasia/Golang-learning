# ✅ARP 与 RARP 的区别是什么？

# 典型回答


ARP（Address Resolution Protocol）和RARP（Reverse Address Resolution Protocol）都是**网络通信协议**，**用于将IP地址和MAC地址进行转换**。它们之间的区别如下：



**功能不同**



ARP协议用于将IP地址转换为MAC地址，也就是在通信时需要知道目标机器的MAC地址时，ARP协议可以用来查询目标机器的MAC地址。



RARP协议则是将MAC地址转换为IP地址，也就是在启动时，需要知道自己的IP地址时，可以向网络中发送RARP请求，获取自己的IP地址。



**工作方式不同**



ARP协议是一种广播协议，当一台主机需要知道另一台主机的MAC地址时，会在本地局域网上广播一个ARP请求包，所有主机都能收到这个请求包，但只有目标主机会响应这个请求，将自己的MAC地址发送回来。

而RARP协议则是向预定义的RARP服务器发出请求，请求服务器返回自己的IP地址。



**用途不同**



由于现代的操作系统和网络设备都可以自动分配IP地址，RARP协议已经很少使用了。而ARP协议则在现代网络通信中仍然起着重要的作用。



> 更新: 2024-09-13 21:31:40  
> 原文: <https://www.yuque.com/hollis666/ukxanr/ckw29kf6luwi1exb>