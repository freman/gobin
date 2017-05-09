onmessage = function(event) {
	importScripts('/static/highlight.pack.js');
	if (event.data.guess === true) {
		var result = self.hljs.highlightAuto(event.data.code, event.data.syntaxes);
		postMessage(result.language);
	} else {
		var result = self.hljs.highlight(event.data.syntax, event.data.code, true);
		postMessage(result.value)
	}
}
