// .commitlintrc.js
/** @type {import('cz-git').UserConfig} */
module.exports = {
  rules: {
    // @see: https://commitlint.js.org/#/reference-rules
  },
  prompt: {
    messages: {
      type: "选择你要提交的类型 | type:",
      scope: "选择一个提交范围（可选）| scope(optional):",
      customScope: "请输入自定义的提交范围 | custom scope:",
      subject: "填写简短精炼的变更描述 | subject:\n",
      body: '填写更加详细的变更描述（可选）。使用 "|" 换行 | body(optional):\n',
      breaking: '列举非兼容性重大的变更（可选）。使用 "|" 换行 | breaking(optional):\n',
      footerPrefixsSelect: "选择关联issue前缀（可选）| footer prefixs(optional):",
      customFooterPrefixs: "输入自定义issue前缀 custom footer prefixs:",
      footer: "列举关联issue (可选) 例如: #31, #I3244 | footer(optional):\n",
      confirmCommit: "是否提交或修改commit | confirm commit?"
    },
    types: [
      {value: 'feat',     name: 'feat:     新增功能 | A new feature'},
      {value: 'fix',      name: 'fix:      修复缺陷 | A bug fix'},
      {value: 'docs',     name: 'docs:     文档更新 | Documentation only changes'},
      {value: 'style',    name: 'style:    代码格式 | Changes that do not affect the meaning of the code'},
      {value: 'refactor', name: 'refactor: 代码重构 | A code change that neither fixes a bug nor adds a feature'},
      {value: 'perf',     name: 'perf:     性能提升 | A code change that improves performance'},
      {value: 'test',     name: 'test:     测试相关 | Adding missing tests or correcting existing tests'},
      {value: 'build',    name: 'build:    构建相关 | Changes that affect the build system or external dependencies'},
      {value: 'ci',       name: 'ci:       持续集成 | Changes to our CI configuration files and scripts'},
      {value: 'revert',   name: 'revert:   回退代码 | Revert to a commit'},
      {value: 'chore',    name: 'chore:    其他修改 | Other changes that do not modify src or test files'},
    ],
    useEmoji: false,
    themeColorCode: "",
    scopes: [
      {name: 'api', value: 'api'},
      {name: 'webhook', value: 'webhook'},
      {name: 'chart', value: 'chart'},
      {name: 'cluster', value: 'cluster'},
      {name: 'backup', value: 'backup'},
      {name: 'user', value: 'user'},
      {name: 'e2e', value: 'e2e'},
      {name: 'github', value: 'github'},
    ],
    allowCustomScopes: true,
    allowEmptyScopes: true,
    customScopesAlign: "bottom",
    customScopesAlias: "custom",
    emptyScopesAlias: "empty",
    upperCaseSubject: false,
    allowBreakingChanges: ['feat', 'fix'],
    breaklineNumber: 100,
    breaklineChar: "|",
    skipQuestions: [],
    issuePrefixs: [
      { value: "Link", name: "Link:     链接关联的 ISSUES"},
      { value: "Fixes", name: "Fixes:   标记 ISSUES 已完成"}
      ],
    customIssuePrefixsAlign: "top",
    emptyIssuePrefixsAlias: "skip",
    customIssuePrefixsAlias: "custom",
    allowCustomIssuePrefixs: true,
    allowEmptyIssuePrefixs: true,
    confirmColorize: true,
    maxHeaderLength: Infinity,
    maxSubjectLength: Infinity,
    minSubjectLength: 0,
    scopeOverrides: undefined,
    defaultBody: "",
    defaultIssues: "",
    defaultScope: "",
    defaultSubject: ""
  }
};
