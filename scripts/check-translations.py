#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Checks translation quality in other language files using zh-CN.json as the baseline.
Rules:
  1. MISSING_KEY: Key is missing.
  2. UNTRANSLATED: Target value exactly matches the Chinese source for non-zh-CN languages.
  3. PLACEHOLDER_MISMATCH: Placeholder or HTML tag set differs after normalization.
Note: Han characters are valid in Japanese and Korean, so only exact string matches are treated as untranslated.
"""
import json
import re
import os

LANG_DIR = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                        "app", "appearance", "langs")
BASE = "zh-CN"
TARGETS = ["ar", "de", "en", "es", "fr", "he", "hi", "id", "it", "ja", "ko",
           "nl", "pl", "pt-BR", "ru", "sk", "th", "tr", "uk", "zh-TW"]

CN_RE = re.compile(r"[\u4e00-\u9fff]")
TRIVIAL_RE = re.compile(r"^[\s\d\W]+$", re.UNICODE)


def load(name):
    with open(os.path.join(LANG_DIR, name + ".json"), encoding="utf-8") as f:
        return json.load(f)


def walk(obj, prefix=""):
    out = {}
    if isinstance(obj, dict):
        for k, v in obj.items():
            key = prefix + "." + k if prefix else k
            if isinstance(v, dict):
                out.update(walk(v, key))
            else:
                out[key] = v
    return out


def norm_placeholders(s):
    """Normalize placeholder comparison by unifying quotes and ignoring help-link URL/domain differences."""
    s2 = s.replace('"', "'")
    s2 = re.sub(r"b3log\.org/siyuan(/[a-z]{2})?/", "scribli.local/", s2)
    s2 = re.sub(r"b3log\.org/siyuan(\?[^\s\"']*)?", "scribli.local", s2)
    s2 = re.sub(r"https://(ld246\.com|liuyun\.io)/article/\d+", "HELPURL", s2)
    ps = re.findall(r"\$\{[^}]+\}|\{[a-zA-Z_][a-zA-Z0-9_]*\}|%[sdf]|\d+\$[sdf]|<[^>]+>", s2)
    return sorted(ps)


def main():
    base = load(BASE)
    bflat = walk(base)
    total = 0
    for t in TARGETS:
        tgt = load(t)
        tflat = walk(tgt)
        issues = []
        for k in bflat:
            if k not in tflat:
                issues.append((k, "MISSING_KEY", ""))
                continue
        for k, bv in bflat.items():
            if k not in tflat:
                continue
            tv = tflat[k]
            if not isinstance(bv, str) or not isinstance(tv, str):
                continue
            if TRIVIAL_RE.match(bv):
                continue
            # Completely untranslated: target value exactly matches the Chinese source for non-zh-CN languages.
            if t != "zh-CN" and tv == bv and CN_RE.search(bv):
                issues.append((k, "UNTRANSLATED", bv[:80]))
            # Placeholder mismatch after normalization.
            bp = norm_placeholders(bv)
            tp = norm_placeholders(tv)
            if bp and bp != tp:
                issues.append((k, "PLACEHOLDER_MISMATCH",
                               "base=%s target=%s" % (bp, tp)))
        if issues:
            total += len(issues)
            print("=" * 70)
            print("### %s  (%d issues)" % (t, len(issues)))
            print("=" * 70)
            for k, kind, detail in issues:
                print("  [%s] %s" % (kind, k))
                if detail:
                    print("      -> %s" % detail[:200])
    print("\nTotal issues: %d" % total)


if __name__ == "__main__":
    main()
