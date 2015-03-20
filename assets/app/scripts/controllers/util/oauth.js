'use strict';

/**
 * @ngdoc function
 * @name openshiftConsole.controller:OAuthController
 * @description
 * # OAuthController
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('OAuthController', function ($location, $q, RedirectLoginService, DataService, AuthService, Logger) {
    var authLogger = Logger.get("auth");

    RedirectLoginService.finish()
    .then(function(data) {
      var token = data.token;
      var then = data.then;
      
      // Redirect to the root if we didn't get a token or an error
      // Typically, this means we accessed /oauth directly, rather than via an auth redirect
      if (!token) {
        authLogger.log("OAuthController, no token or error, redirecting to /");
        $location.url('/');
        return;
      }

      // Try to fetch the user
      var opts = {errorNotification: false, http: {auth: {token: token, triggerLogin: false}}};
      authLogger.log("OAuthController, got token, fetching user", opts);

      DataService.get("users", "~", {}, opts)
      .then(function(user) {
        // Set the new user and token in the auth service
        authLogger.log("OAuthController, got user", user);
        AuthService.setUser(user, token);

        // Redirect to original destination (or default to '/')
        var destination = then || '/';
        if (URI(destination).is('absolute')) {
          if (debug) { Logger.log("OAuthController, invalid absolute redirect", destination); }
          destination = '/';
        }
        authLogger.log("OAuthController, redirecting", destination);
        $location.url(destination);
      })
      .catch(function(rejection) {
        // Handle an API error response fetching the user
        var redirect = URI('/error').query({error: 'user_fetch_failed'}).toString();
        authLogger.error("OAuthController, error fetching user", rejection, "redirecting", redirect);
        $location.url(redirect);
      });

    })
    .catch(function(rejection) {
      var redirect = URI('/error').query({
      	error: rejection.error || "",
      	error_description: rejection.error_description || "",
      	error_uri: rejection.error_uri || "",
      }).toString();
      authLogger.error("OAuthController, error", rejection, "redirecting", redirect);
      $location.url(redirect);
    });

  });
