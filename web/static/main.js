(function() {
	$('.dz').dropzone({
		url: "/f/",
		clickable: false,
		createImageThumbnails: false,
		maxFiles: 1,
		autoProcessQueue: true,
		init: function() {
			this.on("success", function(file, result, myEvent) {
				console.dir(result)
				if (result.response === "ok" && result.target) {
					window.setTimeout(function() { window.location.href = result.target; }, 500);
				} else {
					alert('it broke');
				}
			});
		}
	}).on("click", function() {
		$(this).next().find('input,textarea').first().focus().closest("body").removeClass("noscroll");
		$(this).hide();
	}).closest('body').addClass("noscroll");

	$('#dnd').click(function(e) {
		e.preventDefault();
		$('.dz').show();
	})

	if (hljs) {
		var languages = hljs.listLanguages();

		var delay = (function(){
			var timer = 0;
			return function(callback, ms){
			clearTimeout (timer);
			timer = setTimeout(callback, ms);
			};
		})();

		languages.sort();

		$('input[name="syntax"]')
			.on('focus', function() {
				$(this).removeClass('auto');
			}).autocomplete({
				source:  languages,
				minLength: 0,
				html: false,
				select: function() {
					$('input[name="syntax"]').removeClass('auto');
				}
			}).closest('form').find('.show').click(function(e) {
				e.preventDefault();
				$('input[name="syntax"]').autocomplete("search", "");
			});

		var worker;
		if (typeof(Worker) !== "undefined") {
			worker = new Worker('/static/worker.js');
		}

		$('textarea').on('keyup', function(e) {
			var t = $(this)
			delay(function() {
				if (typeof(worker) !== "undefined") {
					worker.onmessage = function(event) { $('input[name="syntax"].auto').val(event.data); }
					worker.postMessage({syntaxes: GuessLanguages, code: t.val(), guess: true});
				} else {
					$('input[name="syntax"].auto').val(hljs.highlightAuto($(this).val(), guessLanguages).language)
				}
			},500);
		})
	}
})();
