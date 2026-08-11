# Teaching Example 15

~90 min. For an hour, cut step4a and run step5 with two defenses.

One shape: the first half builds four defenses that read the attacker's text and
judge it, the second half four that never read it. Step6 separates them.
Everything before it is setup.

Payloads are not reproduced here - cite the file, let the run print it.

## Setup

```
make compose-up       # postgres + pgvector, 5433
make kronk-up
make example15-matrix # ~30 min; keep the summary table for the closing
```

Start [`make example15-attacker`](../../../makefile#L280) in a second terminal,
visible to the room, before you begin. Without it the first half is just logs.

Each step is `make example15-stepN`. Steps 1 and 4 need `kronk-up` only; 2, 3, 5,
6 also need `compose-up`.

## Step1, reconnaissance (10 min)

[`attacks/payloads.go:17`](attacks/payloads.go#L17) `ReconPrompt` - no rule
broken, nothing to alert on. The model returns its tool block verbatim, schema
included.

Say this before anyone decides the demo is staged: the point is not a leak. Tool
definitions ship as structured data on every request - you can watch them go out
on the wire. A design that depends on the attacker not knowing a tool name has no
defense at all.

## Step2, the attack lands (15 min)

Attacker terminal on screen. Say nothing until the confidential record appears in
it, in the clear.

- [`attacks/corpus.go:30`](attacks/corpus.go#L30) `CleanCorpus` - public content
  and one confidential record in the same store. Ordinary, not contrived.
- [`attacks/corpus.go:51`](attacks/corpus.go#L51) `PoisonedDoc` - opens on-topic
  so similarity keeps it in the retrieval set, written in the register of routine
  policy rather than override. Alignment training handles override far better.
- [`attacks/defenses.go:119`](attacks/defenses.go#L119), step 2 of the workflow -
  **the entire indirect-injection vulnerability**, and it reads like good
  instruction-following in a PR.

2b is the moment, not 2a: the attacker is not in the conversation, they wrote a
row months ago and a legitimate question fires it. Ask who in their systems can
write a row that lands in a retrieval context - tickets, PR bodies, changelogs,
scraped docs, CRM notes.

Nothing is simulated. Real tool call, real socket.
[`harness/agent.go:78`](harness/agent.go#L78) `Ran` reports what the tool did,
not what the model said. Hold onto it: in step5 the same call gets made and goes
nowhere, and the gap between those runs is the session.

## Step3, the content defenses hold (12 min)

They hold. The relief is the point.

- [`attacks/defenses.go:125`](attacks/defenses.go#L125) `AntiInjectionRule` -
  appended, so the rule is the only delta from step2.
- [`attacks/defenses.go:146`](attacks/defenses.go#L146) - eleven regexes, every
  entry a phrase someone already published.
- [`attacks/defenses.go:197`](attacks/defenses.go#L197) `ApplyClassifier` - "if
  unsure, default to YES", every advantage on purpose.

Then the question the session turns on: **what is each keyed on?** The program
puts it on screen and stops there; the answer is printed nowhere. It is: the
blocklist matches bytes, the classifier reads English. Wait them out.

## Step4, a second channel (12 min)

No database, which is the point.

**4a, no model anywhere.** [`attacks/vectors.go:340`](attacks/vectors.go#L340)
`DirectCalls`, [`attacks/tools.go:334`](attacks/tools.go#L334)
`RegisterVulnerableShell` - one free-form string, described as a command line,
handed to a shell. That contract is unsafe for every caller it will ever have:
model, retry loop, queue consumer, test. Nobody had to be tricked. Say this or
"the model is the problem" eats ten minutes.

**4b, injection as tool output.** The user asks about the weather.
[`attacks/tools.go:252`](attacks/tools.go#L252) `InjectedToolOutput` -
compromising the weather *service* was never necessary; any upstream field
populated from user-influenced data will do. Ask which of their systems take
model context from an upstream API like that - the step2 question, one layer out.
The model calls the shell tool on its own. Every step3 filter was inspecting the
user's message, which was a weather question.

**4c, sanitize the tool output.** It works -
[`attacks/defenses.go:252`](attacks/defenses.go#L252) and
[`:238`](attacks/defenses.go#L238). Then name it: the step3 blocklist in a new
position. Relocating a content filter does not change what it is keyed on, a
promise you are making about step6.

## Step5, where the defense sits (15 min)

[`attacks/defenses.go:29`](attacks/defenses.go#L29), the `Defenses` struct on
screen: [content-side](attacks/defenses.go#L30) and
[structural](attacks/defenses.go#L36). The boundary is the lesson.

- **5a, egress.** [`attacks/tools.go:193`](attacks/tools.go#L193). After the
  model decided, after arguments are final, before the socket. Attacker terminal
  stays empty.
- **5b, schema + allowlist.** [`attacks/tools.go:417`](attacks/tools.go#L417) and
  [`:495`](attacks/tools.go#L495). Both exist because a schema constraint is a
  hint the model may ignore; the same constraint must run as code after arguments
  arrive.
- **5c, authorization.** [`attacks/tools.go:464`](attacks/tools.go#L464). Caller
  off the context, unreachable by the model.
- **5d, access level.** [`attacks/defenses.go:325`](attacks/defenses.go#L325) and
  [`harness/harness.go:250`](harness/harness.go#L250). **Slow down here.**
  `NEUTRALIZED`, not `REFUSED`: nothing was caught or blocked. The attack ran end
  to end, the collector returned 200, and what left was public content. One
  argument to the retrieval query.

The honest part: step3 also stopped the plain-English payload. Side by side the
blocklist and the allowlist look equally good. Only "what is it keyed on"
separates them.

Ask how they'd reword to get past 5b. Not to make the model try - 5b already
shows that, and [`attacks/outcome.go:84`](attacks/outcome.go#L84) scores it
`REFUSED` to keep "the model asked" apart from "the tool ran". Ask how to phrase
a request for an operation that isn't in the allowlist. No wording does it; the
check reads no words.

## Step6, the punchline (15 min)

All content-side defenses on, nothing structural. One claim.

The step prints both forms of the payload - read them off the output.
[`attacks/payloads.go:37`](attacks/payloads.go#L37) is the plain form, which the
blocklist catches. Below it, the same intent in Morse, a wall of dots and dashes
attached to an ordinary question. The Morse is a prop, not a working decode: the
point is that not one keyword survives into the text the filters read.

Both verdicts say `BYPASSED` and they do not mean the same thing:

- The input sanitizer matched **zero patterns** - point at its own line saying
  so. It abstained, it did not fail. Hence
  [`attacks/defenses.go:173`](attacks/defenses.go#L173) returning whether it
  redacted anything; an abstention must not print like a success.
- The classifier at [`attacks/defenses.go:197`](attacks/defenses.go#L197) got its
  chance, read a plain question with a Morse attachment, judged it benign, and
  was right about the surface. Wrong thing to be reading.

**The guard is not underpowered.**
[`harness/harness.go:60`](harness/harness.go#L60) - the classifier runs the same
8B that answers. Measured: it **flags** step3's plain-English payload and
**passes** this one. Nobody can write the bypass off as a cheap filter, the
objection a smaller guard invites. Production does put something small in front;
that only widens the window.

Don't dwell on whether the model would decode the Morse. That is not the control
in question - nobody wrote it, nobody reviews it, and it moves on someone else's
release schedule. The only claim the step makes is the one the two `BYPASSED`
verdicts prove: the content-side filters never saw the payload at all.

## Closing, one word (5 min)

Live, in the editor. [`step6/main.go:51`](step6/main.go#L51), one field:

```go
content := attacks.Defenses{
    HardenedPrompt:      true,
    InputSanitizer:      true,
    ToolOutputSanitizer: true,
    Classifier:          true,
    AccessFilter:        true,   // <- the only edit
}
```

Rerun. Byte-identical input, both content filters still `BYPASSED`, outcome
changes anyway. The payload was never what `AccessFilter` reads - it is an
argument to a retrieval query
([`attacks/defenses.go:325`](attacks/defenses.go#L325)), so no rewording reaches
it. Every flag in the second group is keyed on something the attacker does not
author: document metadata, a schema constraint, the caller on the context, a
destination host.

Content-side defenses are worth having, and worth exactly what step6 showed. The
question for any proposed AI security control is not "does it catch this payload"
but **"is it reading something the attacker controls?"**
