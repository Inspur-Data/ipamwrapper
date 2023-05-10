# k8-ipam

## 介绍

k8-ipam是一个 kubernetes 的 IPAM 插件项目， 针对云原生网络的 IP 地址管理需求而设计，并且可以与开源社区中主流开源CNI项目兼容，如Calico、Cilium、kube-ovn等，为其提供IP地址管理功能。当前开源社区中已有部分IPAM项目，典型的如whereabouts、calico-ipam，暂时还没有一个IPAM插件可以满足云原生场景中的所有IPAM需求，例如Pod固定IP功能，虽然Calico-ipam和kube-ovn可以实现该功能，但都通过hard-code方式实现，灵活性较低。k8-ipam旨在解决云原生场景中Pod固定IP、子网、定制化路由、IPv4/IPv6双栈、预留IP等需求，并基于kubernetes CRD进行管理，极大简化IPAM的运维管理工作。

## 关键功能

| 功能          | 描述 |      |
| :------------ | ---- | ---- |
| 应用固定IP    |      |      |
| IPv4/IPv6双栈 |      |      |
| 子网          |      |      |
| 定制化路由    |      |      |
| 预留IP        |      |      |



## 架构

## 快速搭建

参考[快速搭建](./docs/install.md)

## RoadMap

参考[roadmap](roadmap.md)

## 联系我们



 - 我们来自浪潮云海k8s开发团队

 - [roadMap](roadmap.md)

 - 欢迎有想法的小伙伴参与此项目中

   
