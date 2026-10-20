"""Extract the official Greek description of every field into a JSON map."""

import glob
import json
import os
import xml.etree.ElementTree as ET

SRC = r"C:\Users\Takis\AppData\Local\Temp\opencode\ergani-docs\json"
XS = "{http://www.w3.org/2001/XMLSchema}"
OUT = r"C:\Users\Takis\AppData\Local\Temp\opencode\field-docs.json"

docs = {}
for path in sorted(glob.glob(os.path.join(SRC, "*.xsd"))):
    root = ET.parse(path).getroot()
    for el in root.iter(f"{XS}element"):
        field = el.get("name")
        if not field or not field.startswith("f_"):
            continue
        for ann in el.iter(f"{XS}documentation"):
            text = " ".join((ann.text or "").split())
            if text and field not in docs:
                docs[field] = text
            break

with open(OUT, "w", encoding="utf-8") as fh:
    json.dump(docs, fh, ensure_ascii=False, indent=1, sort_keys=True)

print(f"wrote {len(docs)} field descriptions to {OUT}")
