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

	jQuery.fn.extend({
		lineNumbers: function() {
			if (this.find('.line-number').length) {
				return
			}

			fixWrap = function fixWrap() {
				lines = this.find('code span.line');
				foo = this.find('.line-number span.line').each(function (i, v) {
					$(v).attr('style', 'height: ' + $(lines.get(i)).height() + 'px !important;');
				})
			}.bind(this)

			fixWrap();
			$(window).resize(fixWrap);

			this
				.prepend('<span class="line-number hljs"></span>')
				.append('<span class="cl"></span>').each(function() {
					var numbers = $(this).find('.line-number');
					output = '';
					$(this).find('code').html().split(/\n/).forEach(function(v, i) {
						output += '<span class="line" id="l' + (i + 1) + '">' + v + '</span>' + "\n";
						numbers.append('<span class="line" id="c' + (i + 1) + '">' + (i + 1) + '</span>')
					});
					$(this).find('code').html(output);
					window.setTimeout(fixWrap, 150);
				});

			this.find('.line-number').on('click', 'span', function() {
				window.location.hash = $(this).attr('id').replace('c', 'l');
			})

		}
	})

	if (hljs) {
		var languages = hljs.listLanguages();

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

		$('textarea').on('keyup', function(e) {
			$('input[name="syntax"].auto').val(hljs.highlightAuto($(this).val(), guessLanguages).language)
		})
	}
})();
