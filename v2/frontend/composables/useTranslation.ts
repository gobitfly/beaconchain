import type { MessageSchema } from "~/i18n.config";

export function useTranslation() {
  // enables autocompletion
  // https://vue-i18n.intlify.dev/guide/advanced/typescript.html#resource-keys-completion-supporting
  return {
    t: useI18n<{ message: MessageSchema }>({
      useScope: "global",
    }).t,
  };
}
