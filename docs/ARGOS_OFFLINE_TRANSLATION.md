# Offline DOCX translation (optional and experimental)

OpenDesk Free v0.6.8 can call a *locally installed* Argos Translate Python environment. This is an optional feature, not included in the base EXE. It currently works only on simple Word paragraphs with a single w:t text run. It does **not** preserve page wrapping exactly and refuses complex paragraphs rather than damaging formatting.

1. Install Python locally from its official distribution. Make sure `python` is available on PATH (Windows: test `python --version`).
2. Install the open-source Argos Translate package into that same Python environment: `python -m pip install argostranslate`. Review current package dependencies, licenses and origin before installing.
3. Install a direct source→target `.argosmodel` language pack into that Python environment through the official Argos package manager or its local Python API. Model packages are usually large and **are not bundled, verified or automatically downloaded by OpenDesk**.
4. Restart OpenDesk and select **Translate DOCX — Offline beta**. Click **Recheck installed language models**. Only language pairs actually detected on your PC are selectable.
5. Choose one DOCX and process. If FormGuard encounters tracked changes, multiple formatted runs in a paragraph or additional headers/footers, it will BLOCK translation before writing any output.
6. Open the output ZIP: `<name>_<target>.docx`, `Translation_Audit.csv` and `MANIFEST.txt`. Review every translated paragraph and final page layout before any official use.

The app never performs silent pivot translations or calls cloud services. No model authorizations/verified signing are implied by installed packages. Avoid models from untrusted websites. A successful output means only that the Word XML remains syntactically valid and protected placeholders survived; it is NOT a certified translation.

References: https://argos-translate.readthedocs.io/en/stable/ and https://github.com/argosopentech/argos-translate .
