# License and external software policy

OpenDesk Free source is MIT (see /LICENSE). Its pdfcpu dependency is Apache-2.0; confirm transitive dependencies and supply notices where required.

**Current release bundles none of the 44 independent projects listed by the Community Software Directory.** The directory only renders local static text and links; clicking a link is the user's explicit navigation to the upstream project. A GitHub repository being public or software being available free does not grant arbitrary sublicensing.

Before an adapter/embedded binary can move beyond DIRECTORY or ROADMAP:
1. Record source repository, release/tag/version and valid license texts and trademark restrictions.
2. Verify licensing of code modifications, linking, bundling, third-party packages and any software patents if relevant.
3. Check platform, security, download size, offline requirements, source supply/attribution and localization.
4. Build and test it in a CI job (then real Windows and other target OS smoke tests) and show it only under its actual state.
5. Do not use n8n/Dify/Airbyte or commercial enterprise modules as though they are permissively licensed.
6. No automatic installation or download of external executables in the current app.

References: https://www.gnu.org/licenses/license-compatibility.en.html and https://www.libreoffice.org/licenses/
