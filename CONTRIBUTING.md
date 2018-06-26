# Contributing

The Draft project accepts contributions via GitHub pull requests. This document outlines the process to help get your contribution accepted.

## Reporting a Security Issue

Most of the time, when you find a bug in Draft, it should be reported using GitHub issues. However, if you are reporting a security vulnerability, please email a report to one of the [core maintainers][owners] directly. This will give the maintainers a chance to try to fix the issue before it is exploited in the wild.

## Contributor License Agreements

We'd love to accept your patches! Before we can take them, we have to jump a couple of legal hurdles.

Community contributors to code repositories open sourced by Microsoft must sign the [Microsoft Contributor License Agreement][cla]. By signing a CLA, we ensure that the community is free to use your contributions. 

Once you are CLA'ed, we'll be able to accept your pull requests. For any issues that you face during this process, please write a comment explaining the issue and we will help get it sorted out.

When you contribute to a Microsoft open source project on GitHub with a new pull request, a bot will evaluate whether you have signed the CLA. If required, the bot will comment on the pull request, including a link to the CLA system to accept the agreement.

## Pull Requests

Like any good open source project, we use Pull Requests to track code changes.

PRs that are currently in progress are more than welcome. They are a great way to keep track of important work that is in-flight, but useful for others to see. If a PR is a work in progress, it must be prefaced with "WIP: [title]". Once the PR is ready for review, remove "WIP" from the title.

The maintainer in charge of triaging will apply the proper labels for the issue. This should include at least a category label (like `area/docs`), and a bug or feature label. This allows us to more efficiently triage the issue queue.

When reviewing pull requests, Draft uses a LGTM (Looks Good To Me!) policy. Because of the velocity of the project in its given state (pre-v1.0), the LGTM policy is as follows:

### Pull Requests Submitted by Admirals

Small PRs submitted by an Admiral only requires a single LGTM from another Admiral or a Commodore.
This is because an Admiral is identified as an individual with significant experience with the
project, so it is assumed that smaller features have already been "signed off".

Larger PRs that alter behaviour significantly from what's in master needs to be signed off by two
Admirals or Commodores, but only one of them needs to review it. This is to ensure a proper transfer
of knowledge is passed on to other Admirals and Commodores, reducing overall
[bus factor][], while still ensuring the project can
continue at its current velocity.

The sign-off process is completely informal. A "full steam ahead!" on Slack is more than acceptable.

Scenario: there are two Admirals and a Commodore. Admiral "a" proposes a certain feature that alters
how Draft operates in a significant way. Admiral "b" and the Commodore both approve the proposal
(informally), and Admiral "b" reviews the pull request.

### Pull Requests Submitted by Commodores

The same policy applied to Admirals also applies to Commodores. Commodores are seen in the same
light as Admirals when it comes to code contributions; they just have less overall responsibility to
maintain the project's direction and governance.

### Pull Requests Submitted by the Community

All PRs, small or big, need to be signed off by two Admirals/Commodores and reviewed by one.

### Merging PRs

PRs should stay open until merged or if they have not been active for more than 30 days. This will help keep the PR queue to a manageable size and reduce noise. Should the PR need to stay open (like in the case of a WIP), the `wip` label can be added.

If the owner of the PR is listed in OWNERS, that user may merge their own PRs or explicitly request another OWNER do that for them.

If the owner of a PR is not listed in OWNERS, any OWNER may merge the PR once it is approved.

## Code of Conduct

This project has adopted the [Microsoft Open Source Code of Conduct][coc]. For more information see the [Code of Conduct FAQ][coc faq] or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.


[bus factor]: https://en.wikipedia.org/wiki/Bus_factor
[cla]: https://cla.opensource.microsoft.com/
[coc]: https://opensource.microsoft.com/codeofconduct/
[coc faq]: https://opensource.microsoft.com/codeofconduct/faq/
[owners]: OWNERS
[semver]: http://semver.org/
