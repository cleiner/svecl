import sib from "./sib.js";
import lib from "lib";
import ref from "lib/ref";
import a from "sub/a.js";
import b from "sub/b.js";
// NodeJS-style resolution heuristic
import c from "sub/c";
// test normalization for exact bare specifier matches
import { Y } from 'lib';
import { X } from 'lib/ref';

console.log(sib, lib, ref, a, b, c, X, Y);
