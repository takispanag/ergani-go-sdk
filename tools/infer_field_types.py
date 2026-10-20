"""Infer the type of every field of every declaration, from the XSD schemas.

The XSD schemas of the Ministry declare the format of every field, which is
what tells the type of the field in Go: a date, a time, a number of the
Greek format, a count, a flag, a document or a plain string.
"""

import glob
import json
import os
import re
import xml.etree.ElementTree as ET

SRC = r"C:\Users\Takis\AppData\Local\Temp\opencode\ergani-docs\json"
OUT = r"C:\Users\Takis\AppData\Local\Temp\opencode\field-types.json"
XS = "{http://www.w3.org/2001/XMLSchema}"

# The Greek number of the API: the thousands are separated by a dot and the
# decimals by a comma.
GREEK_AMOUNT = re.compile(r"\(,\[0-9\]\[0-9\]\)$")
GREEK_HOURS = re.compile(r"\(,\[0-9\]\)$")
DATE = re.compile(r"\[/\]\)\(\[1-2\]\)\(\[0-9\]\)\(\[0-9\]\)\(\[0-9\]\)")
TIME = re.compile(r"\(\[0-1\]\[0-9\]\|2\[0-3\]\):\(\[0-5\]\[0-9\]\)")

# The type of every field, per schema.
types = {}

for path in sorted(glob.glob(os.path.join(SRC, "*.xsd"))):
    name = os.path.basename(path)[:-4]
    root = ET.parse(path).getroot()

    for element in root.iter(XS + "element"):
        field = element.get("name")
        if not field or not field.startswith("f_"):
            continue

        simple = element.find(XS + "simpleType")
        if simple is None:
            types.setdefault(name, {})[field] = "string"
            continue

        restriction = simple.find(XS + "restriction")
        if restriction is None:
            types.setdefault(name, {})[field] = "string"
            continue

        base = restriction.get("base", "")
        pattern_element = restriction.find(XS + "pattern")
        pattern = pattern_element.get("value") if pattern_element is not None else ""
        enums = [e.get("value") for e in restriction.findall(XS + "enumeration")]

        if base == "xs:base64Binary":
            kind = "base64"
        elif base == "xs:date":
            kind = "date"
        elif base == "xs:dateTime":
            kind = "datetime"
        elif enums and set(enums) <= {"0", "1"}:
            kind = "flag"
        elif enums:
            kind = "enum"
        elif pattern and GREEK_AMOUNT.search(pattern):
            kind = "amount"
        elif pattern and GREEK_HOURS.search(pattern):
            kind = "hours"
        elif pattern and TIME.search(pattern):
            kind = "time"
        elif pattern and DATE.search(pattern):
            kind = "date"
        elif pattern and re.fullmatch(r"\[0-9\]\{1,2\}", pattern):
            kind = "count"
        elif pattern and re.fullmatch(r"\[0-9\]\{1,3\}", pattern):
            kind = "minutes"
        else:
            kind = "string"

        types.setdefault(name, {})[field] = kind

with open(OUT, "w", encoding="utf-8") as fh:
    json.dump(types, fh, ensure_ascii=False, indent=1, sort_keys=True)

kinds = {}
for schema, fields in types.items():
    for field, kind in fields.items():
        kinds[kind] = kinds.get(kind, 0) + 1

print(f"wrote {OUT} with {len(types)} schemas")
for kind, count in sorted(kinds.items(), key=lambda kv: -kv[1]):
    print(f"  {kind:9s} {count}")
