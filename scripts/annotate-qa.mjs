#!/usr/bin/env node
// Annotation QA and repair for swept status reports.
//
// Compares a swept report with a base revision and reports mechanical damage
// the docs gates do not see. With --fix it applies the repairs that are
// purely mechanical; everything else is listed for a human decision.
//
//   node scripts/annotate-qa.mjs [--base REV] [--fix] [--json] <file>...
//
// Findings:
//   R  renumbered: the item's base text sits under another number
//   I  indented: a struck item starts in a column the base does not have
//   S  strike: odd number of ~~ outside code spans in the item
//   X  split: the strike closes before the item's last line
//   U  unclosed: a struck item whose strike never closes
//   A  marker: a routed marker sits before the item's last line
//   W  whitespace: the base's spaces are gone from the item's first line
//   T  text: the item's first line diverges from its base text
//   L  lost: words of the base item are missing
//
// Repairs (--fix): X (strike extended over the continuation lines, the
// verdict moves to the item's end), W (base spaces restored), R (base
// number restored), I (column 0 plus a blank line above).

import { execFileSync } from "node:child_process";
import { readFileSync, writeFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const NUMBERED = /^(\s*)(~~)?(\d+)\.(?:\s+|$)(.*)$/;
const MIN_COMMON_PREFIX = 24;

function parseArgs(argv) {
	const options = { base: "HEAD", fix: false, json: false, files: [] };
	for (let index = 0; index < argv.length; index++) {
		const arg = argv[index];
		if (arg === "--base") options.base = argv[++index];
		else if (arg === "--fix") options.fix = true;
		else if (arg === "--json") options.json = true;
		else if (arg === "-h" || arg === "--help") options.help = true;
		else options.files.push(arg);
	}
	return options;
}

function usage() {
	return readFileSync(fileURLToPath(import.meta.url), "utf8")
		.split("\n")
		.filter((line) => line.startsWith("// ") || line.startsWith("//"))
		.map((line) => line.replace(/^\/\/ ?/, ""))
		.slice(0, 26)
		.join("\n");
}

function stripCode(text) {
	return text.replace(/`[^`]*`/g, "");
}

function strikeCount(text) {
	return (stripCode(text).match(/~~/g) ?? []).length;
}

function isBlockStart(line) {
	return /^\d+\.(?:\s+|$)/.test(line) || /^\s*(#{1,6}\s|<!--)/.test(line);
}

function splitItems(lines) {
	const items = [];
	let current = null;
	for (let index = 0; index < lines.length; index++) {
		const line = lines[index];
		if (line.trim() === "") {
			current = null;
			continue;
		}
		const match = NUMBERED.exec(line);
		if (match && match[1] === "") {
			current = {
				index: items.length,
				start: index,
				end: index,
				indent: match[1],
				struck: Boolean(match[2]),
				number: Number(match[3]),
				text: match[4],
				block: [index],
			};
			items.push(current);
			continue;
		}
		if (current && !isBlockStart(line)) {
			current.block.push(index);
			current.end = index;
		} else {
			current = null;
		}
	}
	return items;
}

function itemLines(item, lines) {
	return item.block.map((index) => lines[index]);
}

function normalize(text) {
	return text.replace(/[~*`\s]/g, "");
}

function closingLine(item, lines) {
	if (!item.struck) return -1;
	const block = itemLines(item, lines);
	let seen = 0;
	for (let offset = 0; offset < block.length; offset++) {
		seen += strikeCount(block[offset]);
		if (seen >= 2) return item.start + offset;
	}
	return -1;
}

function verdictTail(item, lines) {
	const first = lines[item.start];
	const codeFree = stripCode(first);
	const relative = codeFree.indexOf("~~", codeFree.indexOf("~~") + 2);
	return relative < 0 ? "" : first.slice(relative + 2).trim();
}

function matchBase(item, lines, baseIndex) {
	const block = itemLines(item, lines);
	const head = block[0].replace(/^\s*(~~)?\d+\.(?:\s+|$)/, "");
	const headKey = normalize(head);
	const bodyKey = normalize(block.join(""));
	let exact = baseIndex.byHead.get(headKey);
	if (!exact) {
		for (const [key, base] of baseIndex.byHead) {
			if (key.length >= 8 && headKey.startsWith(key)) {
				exact = base;
				break;
			}
		}
	}
	if (exact) {
		return { base: exact, kind: "prefix", missing: missingWords(exact.block, block) };
	}
	let best = null;
	let bestLength = 0;
	for (const base of baseIndex.items) {
		const baseKey = base.key;
		let length = 0;
		const limit = Math.min(baseKey.length, bodyKey.length);
		while (length < limit && baseKey[length] === bodyKey[length]) length++;
		if (length > bestLength) {
			bestLength = length;
			best = base;
		}
	}
	if (best && bestLength >= MIN_COMMON_PREFIX) {
		return { base: best, kind: "diverged", at: bestLength, missing: [] };
	}
	return { base: null, kind: "new", missing: [] };
}

function missingWords(baseBlock, block) {
	const have = new Set(
		block.join(" ").split(/\s+/).map((word) => word.replace(/[~*]/g, "")),
	);
	return baseBlock
		.join(" ")
		.split(/\s+/)
		.filter((word) => word !== "" && !have.has(word.replace(/[~*]/g, "")));
}

function baseWhitespace(cur, ref) {
	const refChars = [...ref];
	const out = [];
	let refIndex = 0;
	let pending = "";
	for (const ch of cur) {
		if (/\s/.test(ch)) {
			pending += ch;
			continue;
		}
		if (refIndex >= refChars.length || refChars[refIndex] !== ch) break;
		const refPending = [];
		while (refIndex > 0 && /\s/.test(refChars[refIndex - 1])) {
			refChars.unshift();
		}
		out.push(pending === "" && refHasSpaceBefore(ref, refIndex) ? " " : pending, ch);
		refIndex++;
		pending = "";
	}
	return null;
}

function refHasSpaceBefore() {
	return false;
}

function reSpace(cur, ref) {
	let result = "";
	let refIndex = 0;
	let pending = "";
	for (const ch of cur) {
		if (/\s/.test(ch)) {
			pending += ch;
			continue;
		}
		if (refIndex >= ref.length || ref[refIndex] !== ch) break;
		const refSpace = refIndex > 0 && /\s/.test(ref[refIndex - 1]);
		result += (pending === "" && refSpace ? " " : pending) + ch;
		refIndex++;
		pending = "";
	}
	result += pending;
	if (refIndex < ref.replace(/\s/g, "").length) return null;
	return result + cur.slice([...cur].findIndex((_, index, all) => index >= 0 && all.slice(0, index).join("").length >= result.length) >= 0 ? 0 : 0);
}

function insertBaseSpaces(cur, ref) {
	let result = "";
	let refIndex = 0;
	let pending = "";
	let consumed = 0;
	for (const ch of cur) {
		consumed++;
		if (/\s/.test(ch)) {
			pending += ch;
			continue;
		}
		if (refIndex >= ref.length || ref[refIndex] !== ch) break;
		const refSpace = refIndex > 0 && /\s/.test(ref[refIndex - 1]);
		result += (pending === "" && refSpace ? " " : pending) + ch;
		refIndex++;
		pending = "";
	}
	return { text: result + cur.slice(consumed - (pending ? [...pending].length : 0)), refIndex };
}

function inspect(lines, baseLines, relPath) {
	const items = splitItems(lines);
	const baseItems = splitItems(baseLines);
	const byHead = new Map();
	for (const base of baseItems) {
		base.block = base.block.map((index) => baseLines[index]);
		base.key = normalize(base.block.join(""));
		byHead.set(normalize(base.block[0].replace(/^\s*(~~)?\d+\.(?:\s+|$)/, "")), base);
	}
	const baseIndex = { items: baseItems, byHead };
	const findings = [];
	const add = (code, item, detail) =>
		findings.push({
			code,
			line: item ? item.start + 1 : 0,
			item: item ? item.number : 0,
			detail,
		});
	for (const item of items) {
		const { base, kind, at, missing } = matchBase(item, lines, baseIndex);
		item.base = base;
		const block = itemLines(item, lines);
		if (kind === "diverged") {
			add("T", item, `diverges from base item ${base.number} after ${at} chars`);
		} else if (kind === "new" && baseItems.length > 0) {
			add("T", item, "no base item matches");
		}
		if (base) {
			if (base.number !== item.number) {
				add("R", item, `base number ${base.number}`);
			}
			if (missing.length > 0) {
				add("L", item, missing.slice(0, 8).join(" ") + (missing.length > 8 ? " ..." : ""));
			}
			if (item.indent !== base.indent) {
				add("I", item, `indent ${item.indent.length}, base ${base.indent.length}`);
			}
			const baseFirst = base.block[0].replace(/^\s*(~~)?\d+\.(?:\s+|$)/, "");
			const curFirst = block[0].replace(/^\s*(~~)?\d+\.(?:\s+|$)/, "");
			const curHead = item.struck ? curFirst.slice(0, stripCode(curFirst).indexOf("~~") >= 0 ? mapToRaw(curFirst, stripCode(curFirst).indexOf("~~")) : curFirst.length) : curFirst;
			if (
				normalize(curHead) === normalize(baseFirst) &&
				curHead !== baseFirst &&
				normalize(curHead).length > 0
			) {
				const fixed = insertBaseSpaces(curHead, baseFirst);
				if (fixed.text !== curHead) {
					add("W", item, `base: ${JSON.stringify(baseFirst.slice(0, 70))}`);
					item.whitespaceFix = fixed.text;
				}
			}
		}
		const strikes = strikeCount(block.join("\n"));
		if (item.struck && strikes % 2 !== 0) add("S", item, `${strikes} strike markers`);
		const close = closingLine(item, lines);
		if (item.struck && close < 0) add("U", item, "strike never closes");
		if (item.struck && close >= 0 && close !== item.end) {
			const later = block.slice(close - item.start + 1).join("\n");
			const repairable =
				strikeCount(lines[item.start]) === 2 && strikeCount(later) === 0 && strikes === 2;
			add("X", item, `${item.end - close} continuation line(s)${repairable ? "" : " (manual)"}`);
			item.splitFix = repairable;
		}
		if (!item.struck) {
			const markerOffset = block.findIndex((line) => /→\s*routed/.test(line));
			if (markerOffset >= 0 && markerOffset !== block.length - 1) {
				add("A", item, "routed marker before the item's last line");
			}
		}
	}
	return { findings, items };
}

function mapToRaw(raw, codeIndex) {
	let count = 0;
	for (let index = 0; index < raw.length; index++) {
		if (raw[index] === "`") {
			const end = raw.indexOf("`", index + 1);
			if (end > index) {
				index = end;
				continue;
			}
		}
		if (count === codeIndex) return index;
		count++;
	}
	return raw.length;
}

function applyFixes(lines, report) {
	let changes = 0;
	const spliceAt = [];
	for (const item of [...report.items].sort((left, right) => right.start - left.start)) {
		const findingCodes = new Set(
			report.findings.filter((finding) => finding.item === item.number && finding.line === item.start + 1).map((finding) => finding.code),
		);
		let first = lines[item.start];
		const match = NUMBERED.exec(first);
		if (!match) continue;
		let number = match[3];
		let text = match[4];
		let indent = match[1];
		if (findingCodes.has("R") && item.base) {
			number = String(item.base.number);
			changes++;
		}
		if (findingCodes.has("I") && item.base && item.base.indent === "") {
			indent = "";
			for (let index = item.start; index <= item.end; index++) {
				lines[index] = lines[index].replace(/^\s+/, "");
			}
			if (item.start > 0 && lines[item.start - 1].trim() !== "") spliceAt.push(item.start);
			changes++;
		}
		if (findingCodes.has("X") && item.splitFix) {
			const codeFree = stripCode(text);
			const close = mapToRaw(text, codeFree.indexOf("~~"));
			const tail = text.slice(close + 2).trim();
			text = text.slice(0, close);
			const last = item.end;
			lines[last] = `${lines[last].trimEnd()}~~ ${tail}`.trimEnd();
			changes++;
		} else if (findingCodes.has("W") && item.whitespaceFix) {
			const struckHead = item.struck ? text.slice(0, mapToRaw(text, stripCode(text).indexOf("~~"))) : text;
			const tail = item.struck ? text.slice(mapToRaw(text, stripCode(text).indexOf("~~"))) : "";
			text = (item.whitespaceFix ?? struckHead) + tail;
			changes++;
		}
		lines[item.start] = `${indent}${match[2] ?? ""}${number}. ${text}`.trimEnd();
	}
	for (const index of spliceAt.sort((left, right) => right - left)) {
		if (lines[index - 1].trim() === "" || lines[index].trim() === "") continue;
		lines.splice(index, 0, "");
		changes++;
	}
	return changes;
}

function baseContent(options, relPath) {
	const spec = options.base.split(":");
	const revision = spec[0] === "HEAD" ? "HEAD" : spec[0];
	try {
		return execFileSync("git", ["show", `${revision}:${relPath}`], {
			encoding: "utf8",
			stdio: ["ignore", "pipe", "ignore"],
		});
	} catch {
		return null;
	}
}
