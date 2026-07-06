/* ePer7shme Jobs — ndërveprimet dhe animacionet */
(function () {
  "use strict";

  var reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  /* ---------- Header: hije kur rrëshqet faqja ---------- */
  var header = document.getElementById("siteHeader");
  var toTop = document.getElementById("toTop");
  function onScroll() {
    var y = window.scrollY;
    if (header) header.classList.toggle("scrolled", y > 24);
    if (toTop) toTop.classList.toggle("show", y > 600);
  }
  window.addEventListener("scroll", onScroll, { passive: true });
  onScroll();

  if (toTop) {
    toTop.addEventListener("click", function () {
      window.scrollTo({ top: 0, behavior: reduceMotion ? "auto" : "smooth" });
    });
  }

  /* ---------- Menuja mobile ---------- */
  var navToggle = document.getElementById("navToggle");
  var mainNav = document.getElementById("mainNav");
  if (navToggle && mainNav) {
    navToggle.addEventListener("click", function () {
      var open = mainNav.classList.toggle("open");
      navToggle.classList.toggle("open", open);
      navToggle.setAttribute("aria-expanded", open ? "true" : "false");
      document.body.style.overflow = open ? "hidden" : "";
    });
    mainNav.querySelectorAll("a").forEach(function (a) {
      a.addEventListener("click", function () {
        mainNav.classList.remove("open");
        navToggle.classList.remove("open");
        navToggle.setAttribute("aria-expanded", "false");
        document.body.style.overflow = "";
      });
    });
  }

  /* ---------- Shfaqja gjatë rrëshqitjes (reveal) ---------- */
  var revealObserver = null;
  if ("IntersectionObserver" in window && !reduceMotion) {
    revealObserver = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          entry.target.classList.add("in");
          revealObserver.unobserve(entry.target);
        }
      });
    }, { threshold: 0.18, rootMargin: "0px 0px -40px 0px" });
  }

  function initReveals(root) {
    var scope = root || document;
    scope.querySelectorAll(".reveal:not(.in)").forEach(function (el) {
      if (revealObserver) revealObserver.observe(el);
      else el.classList.add("in");
    });
    scope.querySelectorAll(".skill:not(.in)").forEach(function (el) {
      if (revealObserver) revealObserver.observe(el);
      else el.classList.add("in");
    });
  }
  initReveals();

  /* ---------- Numëruesit e statistikave ---------- */
  function animateCounter(el) {
    var target = parseInt(el.getAttribute("data-count"), 10) || 0;
    var prefix = el.getAttribute("data-prefix") || "";
    if (reduceMotion) { el.textContent = prefix + target; return; }
    var duration = 1800;
    var start = null;
    function step(ts) {
      if (!start) start = ts;
      var p = Math.min((ts - start) / duration, 1);
      var eased = 1 - Math.pow(1 - p, 4); // easeOutQuart
      el.textContent = prefix + Math.round(eased * target);
      if (p < 1) requestAnimationFrame(step);
    }
    requestAnimationFrame(step);
  }

  if ("IntersectionObserver" in window) {
    var counterObserver = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          animateCounter(entry.target);
          counterObserver.unobserve(entry.target);
        }
      });
    }, { threshold: 0.6 });
    document.querySelectorAll(".stat-num[data-count]").forEach(function (el) {
      counterObserver.observe(el);
    });
  } else {
    document.querySelectorAll(".stat-num[data-count]").forEach(animateCounter);
  }

  /* ---------- Efekti tilt në kartat (vetëm me maus) ---------- */
  var canHover = window.matchMedia("(hover: hover) and (pointer: fine)").matches;
  function addTilt(card) {
    if (card.dataset.tilt) return;
    card.dataset.tilt = "1";
    card.addEventListener("mousemove", function (e) {
      var r = card.getBoundingClientRect();
      var x = (e.clientX - r.left) / r.width - 0.5;
      var y = (e.clientY - r.top) / r.height - 0.5;
      card.style.transform =
        "perspective(900px) rotateX(" + (-y * 6).toFixed(2) + "deg) rotateY(" + (x * 6).toFixed(2) + "deg) translateY(-9px)";
    });
    card.addEventListener("mouseleave", function () {
      card.style.transform = "";
    });
  }
  function initTilt(root) {
    if (!canHover || reduceMotion) return;
    (root || document).querySelectorAll(".service-card, .post-card").forEach(addTilt);
  }
  initTilt();

  /* ---------- Shkëlqimi që ndjek kursorin ---------- */
  var glow = document.querySelector(".cursor-glow");
  if (glow && canHover && !reduceMotion) {
    document.addEventListener("mousemove", function (e) {
      glow.style.left = e.clientX + "px";
      glow.style.top = e.clientY + "px";
    }, { passive: true });
  }

  /* ---------- HTMX: ri-inicializo pas çdo swap ---------- */
  document.body.addEventListener("htmx:afterSettle", function (e) {
    initReveals(e.target.closest ? document : null);
    initTilt();
  });
})();
