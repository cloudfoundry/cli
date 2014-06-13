# README for CF CLI - i18n support

The CloudFoundry (CF) Command Line Interface (CLI) is now ready for internationalization (i18n, for short). This README details what features are available, what are not, how you can contribute, when, as well as the over aching goals we had in mind as we entered this massive update of the CLI.

## What?

The CF CLI is perhaps the most user-facing component of any CF environment. Every user of CF at some point will use the CLI to interact with a specific CF PaaS that they are targeting, be it private or public.

As CF is a global operating system for clouds, it only made sense for it to be accessible to the many countries all over the world, where English, may not be the best language of communication.

The CF CLI i18n effort enabled the CLI so that all strings that are used to communicate with the end-user are ready to be translated in the user's native tongue or in some other language that is close to that native tongue. So a French Canadian user can use the CLI in French (fr_FR) and Portuguese user can use the CLI in Brazilian Portuguese (pt_BR).

## When?

Available today are:

1. A version of the CLI that is enabled for translation. Going forward, all other versions of the CLI will maintain that i18n enablement. This means that any new strings added to the system will follow i18n enablement guideline, e.g., use Go-style templates and use `T()` functions call to load translated strings.

2. Default language is English in the the en_US locale. This means that any user for any locale will either have strings for their locale loaded, if they exist, otherwise will default to English, specifically the en_US locale. So for instance, a Great Britain user with locale en_GB will default to en_US since there are no en_GB translations.

3. A complete version of the French translation in the fr_FR locale. This translation was done by the CLI team who has a native French speaker as well as non-native speakers with some French experience. We welcome fixes to this translation as well as any other languages. See next sections on how to contribute.

4. Defaulting of language and territory when a specific translation for a territory does not exist. So for instance, a French Canadian speaker with fr_CA locale will have the fr_FR translation strings loaded instead of en_US since that is the closest translation strings to their language and locale.

## How can you contribute?

There are three places where you can contribute today:

1. Give it a test drive. Download the latest CLI and try it in your locale. You should always at least see no difference since English is the default locale if your current locale does not have any translations yet. If your locale is any of the French locales, e.g., fr_CA or fr_FR and others, then you should see that all the CLI strings are now in French.

2. Submit PRs for existing languages to fix them. If you are using the CLI in French or default English and you see typos or grammar errors or ambiguous strings then please report this as issue or better find the appropriate string in the `cf/i18n/resources/<en or fr>/**/<locale>.all.json` and submit a PR with the fix(es).

3. Submit a PR for a new locale. If you are in one of the locales listed in the next section and the status of translation is not Complete and would like to contribute translations for that locale, then please see the initial strings in the file: `cf/i18n/resources/<language>/**/<locale>.all.json` and submit a PR with the `translation` section of the JSON files that comprise the string in the translated locale.

```
[
  {
    "id":"Create an org",
    "translation":"Créez un org"
  },
  ...
]
```

It is very important that you DO NOT change the `id` sections of the JSON files---simply said, do not change any of the `id` strings, but only the `translation` strings. See the French locale translations as examples.

Finally, it is also important not to translate the argument names in templated strings. Templated strings are the ones which contain arguments, e.g., `{{.Name}}` or `{{.Username}}` and so on. The arguments can move to a different location on the translated string, however, the arguments cannot change, should not be translated, and should not be removed or new ones added. So for instance, the following string is translated in French as follows:

```
[
  ...,
  {
    "id": "Creating quota {{.QuotaName}} as {{.Username}}...",
    "translation": "Créez quota {{.QuotaName}} étant {{.Username}}..."
  },
  ...
]
```

## Goals

Our simple goal is to have translations for the top-tier languages first.  This means accepting contributions (PRs or pull requests) for existing translations (fixes) as well as new translations for the remaining top-tier languages for which we do not yet have translations. The current list of the top-tier languages, default territory, and status of translation follows.

Language	            Locale	  Status
--------              ------  --------
English               en_US   Complete
French                fr_FR   Complete
Spanish               es_ES   English files ready (*)
German                de_DE   English files ready
Italian               it_IT   English files ready
Japanese              ja_JA   English files ready
Korean                ko_KO   English files ready
Portuguese (Brazil)   pt_BR   English files ready
Chinese (simplified)  zh_CN   English files ready
Chinese (traditional) zh_HK   English files ready
Russian               ru_RU   English files ready

If you are interested in submitting translations for a non-top-tier locale, then please communicate with us via VCAP-DEV mailing list and we will see what we can do.

(*) We note "English files ready" to mean that our github project contains English versions for the translation files. We are ready for PRs on these files with actual translated strings.
