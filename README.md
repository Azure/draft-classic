![Draft Logo][logo]
[![Build Status](https://circleci.com/gh/Azure/draft.svg?style=svg)](https://circleci.com/gh/Azure/draft)

# Draft: Streamlined Kubernetes Development

(NOTE: insert gif here on a `draft create/draft up`)

Draft makes it easier to build better cloud-native apps more quickly and with less code.

Draft is a high-level framework that encourages rapid development and clean, pragmatic design. Built with love by developers, it takes care of much of the hassle of cloud-native development, so you can focus on writing your app without needing to reinvent the wheel. It’s also free and open source.

Draft takes an opinionated approach to developing applications for [Kubernetes][], building upon tools like [Docker][], [Helm][], and the [Helm Chart format][charts].

## Getting started with Draft

Depending on how new you are to Draft, you can try [a tutorial][tutorials], or just dive straight into the [documentation][docs].

Want to learn more about the inner workings of Draft? Read the [Draft Enhancement Proposals (DEPs)][deps] for a closer look at the architecture and concepts of Draft, and to see whether Draft is right for your project.

We're always looks for contributions in the form of issues, pull requests, and docs changes. If you see anything that would make Draft a better experience for newcomers or yourself, please feel free to [contribute][contributing]!

## Reporting a Security Issue

Most of the time, when you find a bug in Draft, it should be reported using GitHub issues. However, if you are reporting a security vulnerability, please email a report to one of the [core maintainers][owners] directly. This will give the maintainers a chance to try to fix the issue before it is exploited in the wild.

## Support Channels

Whether you are a user or contributor, official support channels include:

- GitHub issues
- Kubernetes Slack in [#draft-dev][] for development-related discussion
- Kubernetes Slack in [#draft-users][] for quick questions, troubleshooting and general discussion

Before opening a new issue or submitting a new pull request, it’s helpful to do a quick search - it’s likely that another user has already reported the issue you’re facing, or it’s a known issue that we’re already aware of.

## Code of Conduct

This project has adopted the [Microsoft Open Source Code of Conduct][conduct]. For more information see the [Code of Conduct FAQ][conduct-faq] or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## License

This software is covered under the MIT license. You can read the license [here][license].

This software contains code from [Helm][], which is covered by the Apache v2.0 license.

You can read third-party software licenses [here][notice].


[contributing]: docs/contributing/README.md
[deps]: docs/reference/README.md
[docs]: docs/README.md
[hacking]: docs/contributing/hacking.md
[license]: LICENSE
[logo]: docs/img/draft-logo.png
[notice]: NOTICE
[owners]: OWNERS
[tutorials]: docs/tutorials/README.md

[#draft-dev]: https://kubernetes.slack.com/messages/C5NNNFB8S/
[#draft-users]: https://kubernetes.slack.com/messages/C5N5YUSSD/
[charts]: https://docs.helm.sh/developing_charts/
[conduct]: https://opensource.microsoft.com/codeofconduct/
[conduct-faq]: https://opensource.microsoft.com/codeofconduct/faq/
[docker]: https://www.docker.com
[helm]: https://helm.sh
[kubernetes]: https://kubernetes.io/
