We are delighted to present version 1.11.0 of Sesame, our layer 7 HTTP reverse proxy for Kubernetes clusters.

There's been a bunch of great contributions from our community for this release, thanks to everyone!

## Sesame-Operator
The new Sesame Operator provides a method for packaging, deploying, and managing Sesame. The operator extends the functionality of the Kubernetes API to create, configure, and manage instances of Sesame on behalf of users. It builds upon the basic Kubernetes resource and controller concepts, but includes domain-specific knowledge to automate the entire lifecycle of Sesame. 

Visit the [getting started guide](https://projectsesame.io/getting-started/#option-2-install-using-operator) on how to quickly get up and running with the operator.

For more information, see the [Sesame operator](https://github.com/projectsesame/sesame-operator) repo.

## Global TLS minimum to 1.2
The default global minimum TLS version is moved to 1.2 from 1.1.
This forces all HTTPProxies and Ingresses to use at least 1.2.

https://github.com/projectsesame/sesame/pull/3112

## Envoy v1.16.2

Sesame supports Envoy v1.16.2 which resolves various CVEs found in Envoy, please upgrade your clusters!

## Envoy XDS Resource Version Support

As mentioned in [Sesame 1.10](https://projectsesame.io/Sesame_v1100/#envoy-xds-v3-support) the `v2` XDS resource version has been removed from Sesame ahead of its removal from Envoy. Please see the [XDS Migration Guide](https://projectsesame.io/guides/xds-migration/) for upgrading your instances of Envoy/Sesame.

__Note:__ This change applies also to any External Auth servers that may be integrated.

## Trigger rebuild for configured secrets

If client certificates, represented in Kubernetes secrets, were changes, Sesame did not notice that change and blocked a valid cert rotation path for users. Sesame v1.11 adds secret references from the configuration file to the list of secrets that will trigger DAG rebuild.  Previously only secrets referred by HTTPProxy and Ingress resources were considered.  The result was that secrets were not picked up correctly if they were created after the creation of HTTPProxy or Ingress themselves triggered a rebuild.

https://github.com/projectsesame/sesame/pull/3191

Thanks to @tsaarni  for the fix and @Zsolt-LazarZsolt for reporting!

## Deprecation Notices
⚠️ Sesame annotations starting with `Sesame.heptio.com` have been removed from documentation for some time. Sesame 1.8 marks the official deprecation of these annotations and have been removed in Sesame v1.11.0.

## Upgrading
Please consult the upgrade [documentation](https://projectsesame.io/resources/upgrading/).

## Community Thanks!
We’re immensely grateful for all the community contributions that help make Sesame even better! For version 1.11, special thanks go out to the following contributors:
- [@invidian](https://github.com/invidian)
- [@alexbrand](https://github.com/alexbrand)
- [@danehans](https://github.com/danehans)
- [@shuuji3](https://github.com/shuuji3)
- [@yoitsro](https://github.com/yoitsro)
- [@bascht](https://github.com/bascht)
- [@tsaarni](https://github.com/tsaarni)
- [@georgegoh](https://github.com/georgegoh)

## Are you a Sesame user? We would love to know!
If you're using Sesame and want to add your organization to our adopters list, please visit this [page](https://github.com/projectsesame/sesame/blob/master/ADOPTERS.md). If you prefer to keep your organization name anonymous but still give us feedback into your usage and scenarios for Sesame, please post on this [GitHub thread](https://github.com/projectsesame/sesame/issues/1269).
