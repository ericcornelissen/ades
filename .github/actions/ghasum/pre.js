// SPDX-License-Identifier: Apache-2.0

import { spawnSync } from "node:child_process";
import * as console from "node:console";
import { platform } from "node:os";
import { join } from "node:path";
import { env, exit } from "node:process";

// --- Context -----------------------------------------------------------------
const OS = platform().toLowerCase();

const JOB = env.GITHUB_JOB;
const SHA = env.GITHUB_WORKFLOW_SHA;
const OWNER = env.GITHUB_REPOSITORY.split("/").at(0);
const PROJECT = env.GITHUB_REPOSITORY.split("/").at(1);
const WORKFLOW = env.INPUT_WORKFLOW.split(/[/@]/g).slice(2,5).join("/");
const GITHUB_TOKEN = env.INPUT_TOKEN;

let cache;
switch (OS) {
case "darwin": cache = "/Users/runner/work/_actions"; break;
case "linux":  cache = "/home/runner/work/_actions";  break;
case "win32":  cache = "D:\\a\\_actions";             break;
}

// --- Main --------------------------------------------------------------------
try {
	exec(
		["go", "run", "github.com/chains-project/ghasum/cmd/ghasum", "verify", "-cache", cache, "-no-evict", "-offline", `${WORKFLOW}:${JOB}`],
		{ cwd: join(cache, OWNER, PROJECT, SHA) },
	);
	exit(0);
} catch (error) {
	console.error(`::error::${error.message}`);
	nuke();
	exit(1);
}

// --- Functions ---------------------------------------------------------------
function exec(command, opts) {
	const cmd = command[0];
	const args = command.slice(1, command.length);
	const { status } = spawnSync(cmd, args, {
		stdio: "inherit" ,
		...opts
	});

	if (status !== 0) {
		throw new Error("Command failed");
	}
}

function nuke() {
	exec(["rm", "-rf", cache]);
}
