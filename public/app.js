const app = document.getElementById("app");
const toastRegion = document.getElementById("toast-region");
const modalRoot = document.getElementById("modal-root");

const translations = {
  zh: {
    navHome: "首页", navWorks: "作品", navAssets: "资产", navTransfer: "资产行权", navAbout: "关于",
    connect: "连接钱包", menu: "菜单", loading: "正在载入", footer: "演示环境 · 不构成数字资产交易或投资建议",
    themeLight: "切换亮色模式", themeDark: "切换暗色模式",
    homeEyebrow: "Clipli 作品发行工作台", homeTitle: "作品有出处，<br>发行有记录",
    homeLead: "把作品内容、HAPW 权益、CLIP 费用和区域发行放到同一工作台。创作者决定边界，发行团队执行，持有人随时核对。",
    browseWorks: "浏览作品", manageAssets: "管理资产", authorized: "已确认授权", overseas: "海外发行作品", visibility: "区域授权状态可查",
    feature1: "作品与 HAPW 同源", feature1Desc: "从作品内容直接进入 HAPW 资产、版权和发行状态。",
    feature2: "区域授权有边界", feature2Desc: "明确发行范围、有效期和权益，提交前完整确认。",
    feature3: "CLIP 账本透明", feature3Desc: "查看积分余额、DEX 购入和授权手续费记录。",
    releaseTitle: "三步开启区域发行", step1: "建立身份", step1Desc: "连接钱包并绑定发行账号。",
    step2: "选择范围", step2Desc: "选择资产、发行区域与授权期限。", step3: "提交发行", step3Desc: "确认条款并创建可追踪的发行授权。",
    deskStatus: "发行台正在更新", signal1: "HAPW #2048 已进入东南亚字幕准备", signal2: "《夜航者》等待权属确认", signal3: "本周新增 12 条区域授权", signal4: "CLIP 服务费记录已同步",
    selectedWorks: "正在推进的作品", selectedLead: "每个作品条目都连接具体的 HAPW、权利范围和发行入口。", viewAllWorks: "查看全部作品",
    practiceEyebrow: "从作品到发行", practiceTitle: "一件作品进入平台后，具体发生什么", practiceLead: "先记录内容和权属，再讨论区域、期限与渠道。每一步都留下可复核的状态。",
    practice1: "作品登记", practice1Desc: "保留创作者、版本、媒介形式和版权说明。", practice2: "关联 HAPW", practice2Desc: "把可持有、可授权的作品权益与内容对应起来。", practice3: "区域执行", practice3Desc: "字幕、配音、渠道上架按授权范围推进。",
    homeBoundaryTitle: "平台负责记录，不替代创作者做决定", homeBoundaryText: "Clipli 不自动扩大授权，也不把 CLIP 余额当成作品所有权。作品边界由权利人确认，平台只负责把操作、费用和结果整理清楚。", readAbout: "了解平台边界", homeClipLink: "查看 CLIP 机制",
    pulseTitle: "发行状态正在变化", pulseDesc: "用三个公开指标看当前工作台的节奏。", pulseAuthorized: "已确认授权", pulseOverseas: "海外发行作品", pulseVisibility: "状态可查比例", activityTitle: "三条账本同时更新", activityDesc: "作品、HAPW 和 CLIP 的变化在同一个时间线上留下位置。", activityWorks: "作品条目", activityAssets: "HAPW 资产", activityLedger: "CLIP 流水",
    worksEyebrow: "Clipli Screening", worksTitle: "作品展示", worksLead: "以 HAPW 资产为线索，浏览来自 Clipli 的短视频与短剧。",
    all: "全部", featured: "精选", play: "播放", views: "播放", close: "关闭", seedSource: "演示素材", details: "查看详情", detailEyebrow: "Work record", detailLead: "查看作品简介、关联 HAPW、发行状态与国内平台入口。", linkedHapw: "关联 HAPW", format: "作品形式", domesticRelease: "国内发行", haiwenPlatform: "海文发平台", openHaiwen: "打开海文发", backWorks: "返回作品",
    assetsEyebrow: "Asset console", assetsTitle: "资产仪表盘", assetsLead: "把 HAPW 作品资产与 CLIP 平台积分放在同一张可核对的账本中。",
    hapwHoldings: "HAPW 资产持仓", hapwLead: "代表 Clipli 平台上的作品与版权权益。", clipBalance: "CLIP 余额", clipLead: "Clipli 平台发行的积分，不设总量，通过外部 DEX 购买。", clipSupply: "供应策略", clipBuy: "获取方式", dexButton: "前往 DEX 兑换", dexNote: "将打开外部 Uniswap 兑换页面；请在钱包确认网络与代币信息。", mechanismDetails: "机制细节",
    holdings: "持仓数量", transferable: "可转移资产", recent: "最近记录", totalValue: "总参考价值",
    convertAsset: "创建发行授权", transferAsset: "发起资产行权", recentTransfers: "最近资产行权",
    asset: "资产", direction: "行权去向", value: "参考价值", status: "状态", date: "时间", clipHistory: "CLIP 过往交易", txType: "交易类型", txAsset: "关联对象", amount: "数量", txHash: "交易标识", positive: "收入", negative: "支出",
    convertEyebrow: "Asset conversion", convertTitle: "把资产转换为发行授权", convertLead: "转换仅授予选定区域和渠道的发行使用权，不转移资产所有权。",
    selectAsset: "选择资产", config: "授权配置", region: "发行范围", duration: "授权期限",
    regionSea: "东南亚地区字幕、本地配音与渠道发行", regionEurope: "欧洲地区字幕与流媒体发行", regionGlobal: "全球数字渠道（不含独家权）",
    day: "天", feeNotice: "确认后将消耗 18 CLIP 创建区域发行授权。资产所有权不变。",
    acceptLicense: "我已阅读并同意本次资产授权范围与商业使用条款。", submitConversion: "确认转换并创建授权",
    conversionDone: "授权请求已提交", unavailable: "当前不可用",
    transferEyebrow: "HAPW ASSET EXERCISE", transferTitle: "HAPW 资产行权", transferLead: "选择 HAPW 与第三方平台，并创建钱包签名请求。",
    outbound: "Clipli → 海文发", inbound: "海文发 → Clipli", sourceAccount: "转出账户", targetAccount: "转入账户",
    transferDetails: "资产行权明细", held: "持有数量", unit: "件", acceptOwnership: "我确认拥有该 HAPW 的行权资格，并同意钱包签名请求。",
    transferNotice: "最终状态以钱包签名、链上记录和第三方平台规则为准。", submitTransfer: "确认行权并请求签名", transferDone: "行权请求已创建，等待钱包签名",
    bindEyebrow: "HWF connect", bindTitle: "绑定海文发账号", bindLead: "同步发行等级、渠道权益与可用服务，不会同步支付密码或身份凭证。",
    overseasAccount: "海文发账号", authorizedAccount: "已授权", phone: "手机号绑定", points: "积分", benefits: "已激活海外发行权益",
    priority: "海外渠道优先级 +1", discount: "服务手续费减免 10%", quota: "每月本地化额度 +60 CLIP",
    bindAction: "绑定账号", unbind: "解除绑定", accountPlaceholder: "输入海文发账号",
    walletEyebrow: "Web3 login", walletTitle: "连接钱包", walletLead: "选择钱包完成演示登录。平台不会索取或保存私钥与助记词。",
    browserWallet: "浏览器扩展钱包", mobileWallet: "移动端连接", walletService: "Web3 钱包服务", selected: "已选择",
    connectSelected: "连接所选钱包", connected: "已连接", chooseWallet: "请先选择钱包",
    securityEyebrow: "Security center", securityTitle: "安全与授权设置", securityLead: "管理签名确认、授权提醒与已连接的生态账号。",
    walletSign: "钱包签名确认", walletSignDesc: "资产转换与转移始终需要钱包签名。", expiryReminder: "授权到期提醒",
    expiryReminderDesc: "授权结束前 7 天发送站内提醒。", justNow: "刚刚", enabled: "已开启", disabled: "已关闭",
    aboutEyebrow: "About Clipli", aboutTitle: "把作品、权益和发行放回同一条链路", aboutLead: "Clipli 面向创作者、内容持有人和区域发行团队，提供一套可理解、可复核、可继续扩展的作品资产工作台。",
    utilityTitle: "HAPW 是作品权益的载体", utilityText: "HAPW 代表 Clipli 平台上的作品资产。它关联作品、角色、版权元数据和可授权范围，持有人可以围绕具体作品发起区域发行，而不是面对一串难以理解的编号。",
    rightsTitle: "CLIP 是平台积分", rightsText: "CLIP 由 Clipli 平台发行，不设总量，主要用于授权手续费、创作激励和服务兑换。用户可通过外部 DEX 购买，余额与每一次使用都会在仪表盘中留下记录。",
    platformRoles: "平台如何协作", creatorRole: "内容创作者", creatorRoleDesc: "发布作品、补充版权资料，决定可发行区域，并在版本变化时更新说明。", distributorRole: "区域发行团队", distributorRoleDesc: "完成字幕、配音、渠道上架和本地化运营，只在明确的授权范围内使用作品。", holderRole: "HAPW 持有人", holderRoleDesc: "核对资产状态，发起授权或转移，同时把 CLIP 费用与作品权益分开管理。", governanceTitle: "一套务实的发行规则", governanceText: "平台先把作品内容、资产状态和区域许可拆开管理，再通过明确的操作记录把它们重新连接。后续 API、钱包或链上服务也必须保留相同字段与审计边界。",
    aboutWhyTitle: "为什么需要 Clipli", aboutWhyText1: "作品进入不同市场时，往往同时出现版本、版权、字幕、配音和渠道问题。信息散在聊天记录、表格和钱包地址里，很难判断谁在什么时间获得了什么权利。", aboutWhyText2: "Clipli 做的事情很朴素：把作品、权益和执行记录放在同一个上下文中，让参与者先看清边界，再继续工作。",
    recordTitle: "平台具体记录什么", recordIntro: "不是给作品贴一个新标签，而是保存之后仍能被核对的事实。", record1Title: "作品版本", record1Desc: "标题、媒介形式、创作者、内容摘要与更新记录。", record2Title: "权益范围", record2Desc: "关联 HAPW、持有人、可发行区域、渠道和有效期。", record3Title: "执行结果", record3Desc: "CLIP 费用、授权申请、转移状态与国内外平台入口。",
    boundaryTitle: "明确不做的事", boundaryText: "平台不托管私钥，不保证 CLIP 价格，不把 HAPW 自动转换成金融收益，也不替代正式版权合同。演示数据只用于说明产品流程。", principleTitle: "工作原则", principle1: "先确认权属，再开始发行", principle2: "每个区域单独描述范围", principle3: "费用与作品资产分账记录", principle4: "外部平台入口保持可追溯",
    aboutWorkflowTitle: "从首次登记到区域发行", aboutWorkflowLead: "一件作品可以有多个区域版本，但每个版本都从同一份权利说明开始。", aboutWorkflow1Title: "先让作品可读", aboutWorkflow1Desc: "记录标题、创作者、媒介、版本和权利备注，形成可复核的作品条目。", aboutWorkflow2Title: "再拆开权益", aboutWorkflow2Desc: "把可持有的 HAPW、可发行区域和授权期限分别写清楚。", aboutWorkflow3Title: "最后交给发行", aboutWorkflow3Desc: "字幕、配音、渠道上架和下架，都回到授权记录中，不依赖口头约定。",
    clipPageEyebrow: "CLIP MECHANISM · V1", clipPageTitle: "CLIP 如何在 Clipli 中流转", clipPageLead: "CLIP 是平台服务的计价与结算单位。它可以从外部 DEX 获得，也可作为创作和发行任务奖励，但不代表作品所有权。", backAssets: "返回资产仪表盘", clipCurrentBalance: "当前可用余额", clipNoPeg: "不与 USDT 固定锚定", clipContractPending: "正式合约信息待发布",
    clipMechanismTitle: "一条简单的使用循环", clipFlowBuy: "USDT 兑换", clipFlowBuyDesc: "用户在外部 DEX 按市场价格获得 CLIP。", clipFlowUse: "支付服务", clipFlowUseDesc: "用于发行授权、本地化和其他平台服务。", clipFlowReward: "任务奖励", clipFlowRewardDesc: "平台按已完成的创作与发行任务发放。", clipFlowRecord: "账本记录", clipFlowRecordDesc: "每一笔收入和支出都回到用户账本。",
    clipRelationTitle: "HAPW 与 CLIP 分别解决什么", hapwMechanismTitle: "HAPW：作品权益", hapwMechanismText: "对应一件作品或一组版权权益，记录持有人、可授权范围和流转状态。", clipMechanismRoleTitle: "CLIP：服务结算", clipMechanismRoleText: "用于支付平台服务或接收任务奖励，不附带对任何 HAPW 的所有权。", clipNoAutoTitle: "两者不会自动兑换", clipNoAutoText: "创建 HAPW 不会自动增发 CLIP，持有 CLIP 也不能直接获得作品权益。两套账本只在授权费用等业务动作中关联。",
    clipExchangeTitle: "使用 USDT 兑换 CLIP", clipExchangeLead: "兑换发生在外部 DEX。Clipli 只提供入口和操作说明，不托管资金，也不承诺固定汇率。", exchange1: "连接支持目标网络的钱包", exchange2: "选择 USDT 作为支付资产", exchange3: "核对正式 CLIP 合约地址和滑点", exchange4: "确认兑换，并等待链上结果", marketPrice: "价格由流动性池决定", marketPriceDesc: "CLIP 不是稳定币，CLIP/USDT 比例会随市场流动性变化。",
    clipPolicyTitle: "机制边界", supplyPolicyTitle: "发行", supplyPolicyText: "不设固定总量。平台仅根据真实服务需求、已完成任务和激励预算释放 CLIP，并保留发行记录。", usagePolicyTitle: "用途", usagePolicyText: "当前用于授权手续费、创作激励、本地化服务和平台服务兑换。", recyclePolicyTitle: "回流", recyclePolicyText: "服务费进入平台服务账户，可用于后续运营与任务奖励；不承诺销毁或固定回购。", disclosurePolicyTitle: "披露", disclosurePolicyText: "正式上线前应公开网络、合约地址、流动性池和发行记录。", clipDemoNotice: "当前为功能演示。正式 CLIP 合约与网络尚未配置，请勿仅凭页面名称进行任何兑换。",
    clipNumbersTitle: "一组可核对的演示数字", clipNumbersLead: "这些数字来自当前演示账本，用来说明 CLIP 会在哪些动作中发生变化。", clipMetricBalance: "当前余额", clipMetricLicense: "单次授权示例", clipMetricLocalization: "本地化示例", clipMetricReward: "创作奖励示例", clipMetricRecords: "历史流水条数", clipMetricQuote: "外部报价资产",
    clipPlayTitle: "四种常见使用场景", clipPlayLead: "把 CLIP 看作平台服务的操作额度，而不是 HAPW 的替代品。", clipPlay1Title: "区域授权", clipPlay1Desc: "选择一个 HAPW、区域和 30/90/180 天期限；当前演示每次消耗 18 CLIP。", clipPlay2Title: "本地化服务", clipPlay2Desc: "字幕、配音和渠道准备可以单独计费；当前账本展示一笔 60 CLIP 示例。", clipPlay3Title: "创作激励", clipPlay3Desc: "已完成的创作或发行任务可以获得 CLIP；当前账本展示一笔 240 CLIP 示例。", clipPlay4Title: "持有与兑换", clipPlay4Desc: "余额可以继续用于服务或在外部 DEX 兑换，兑换价格随流动性变化，不产生 HAPW 所有权。",
    success: "操作成功", failed: "操作失败", noData: "暂无数据", yuan: "¥"
  },
  en: {
    navHome: "Home", navWorks: "Works", navAssets: "Assets", navTransfer: "Asset exercise", navAbout: "About",
    connect: "Connect wallet", menu: "Menu", loading: "Loading", footer: "Demo environment · Not financial or trading advice",
    themeLight: "Switch to light mode", themeDark: "Switch to dark mode",
    homeEyebrow: "Clipli release workbench", homeTitle: "Clear provenance.<br>Traceable releases.",
    homeLead: "Keep work content, HAPW rights, CLIP fees and regional release in one workbench. Creators set the boundary, release teams execute, and holders can verify every step.",
    browseWorks: "Browse works", manageAssets: "Manage assets", authorized: "Verified licenses", overseas: "Overseas releases", visibility: "Visible regional status",
    feature1: "Works meet HAPW", feature1Desc: "Move from a story directly to its HAPW asset, rights and release status.",
    feature2: "Licenses have boundaries", feature2Desc: "Review region, duration and rights before every submission.",
    feature3: "A clear CLIP ledger", feature3Desc: "Track points, DEX purchases and license fees in one history.",
    releaseTitle: "Start a regional release in three steps", step1: "Establish identity", step1Desc: "Connect a wallet and bind a distribution account.",
    step2: "Choose scope", step2Desc: "Select an asset, release region and license term.", step3: "Submit release", step3Desc: "Accept the terms and create a traceable license.",
    deskStatus: "Release desk updating", signal1: "HAPW #2048 entered SEA subtitle prep", signal2: "Night Flyers awaits rights confirmation", signal3: "12 regional licenses added this week", signal4: "CLIP service fees synchronized",
    selectedWorks: "Works in progress", selectedLead: "Each work connects to a specific HAPW, rights scope and release entry.", viewAllWorks: "View all works",
    practiceEyebrow: "From work to release", practiceTitle: "What actually happens after a work enters Clipli", practiceLead: "Content and ownership come first. Region, term and channel follow, with a reviewable state at every step.",
    practice1: "Register the work", practice1Desc: "Keep creator, version, medium and rights notes together.", practice2: "Link HAPW", practice2Desc: "Connect holdable and licensable rights to the work itself.", practice3: "Run the region", practice3Desc: "Subtitles, dubbing and channel listing stay inside scope.",
    homeBoundaryTitle: "Clipli keeps records. Creators keep decisions.", homeBoundaryText: "Clipli does not expand a license automatically or treat a CLIP balance as ownership. Rights holders confirm the boundary; the platform keeps operations, fees and outcomes legible.", readAbout: "Read platform boundaries", homeClipLink: "Read CLIP mechanism",
    pulseTitle: "Release activity in motion", pulseDesc: "Three public signals show the pace of the current workbench.", pulseAuthorized: "Verified licenses", pulseOverseas: "Overseas releases", pulseVisibility: "Visible status", activityTitle: "Three ledgers, one timeline", activityDesc: "Changes to works, HAPW and CLIP keep a place in the same operational timeline.", activityWorks: "Work entries", activityAssets: "HAPW assets", activityLedger: "CLIP ledger",
    worksEyebrow: "Clipli Screening", worksTitle: "Works", worksLead: "Explore Clipli shorts and series through their linked HAPW assets.",
    all: "All", featured: "Featured", play: "Play", views: "views", close: "Close", seedSource: "Seed content", details: "View details", detailEyebrow: "Work record", detailLead: "Review the work, linked HAPW, release state and domestic platform entry.", linkedHapw: "Linked HAPW", format: "Format", domesticRelease: "Domestic release", haiwenPlatform: "HAIWEN platform", openHaiwen: "Open HAIWEN", backWorks: "Back to works",
    assetsEyebrow: "Asset console", assetsTitle: "Asset dashboard", assetsLead: "Keep HAPW work assets and CLIP platform points in one auditable ledger.",
    hapwHoldings: "HAPW holdings", hapwLead: "Work and rights assets issued on Clipli.", clipBalance: "CLIP balance", clipLead: "Points issued by Clipli with no fixed cap, purchased through an external DEX.", clipSupply: "Supply policy", clipBuy: "Acquisition", dexButton: "Swap on DEX", dexNote: "Opens the external Uniswap swap page. Confirm network and token details in your wallet.", mechanismDetails: "Mechanism details",
    holdings: "Holdings", transferable: "Transferable", recent: "Recent records", totalValue: "Reference value",
    convertAsset: "Create release license", transferAsset: "Exercise asset", recentTransfers: "Recent asset exercises",
    asset: "Asset", direction: "Direction", value: "Reference value", status: "Status", date: "Date", clipHistory: "CLIP transaction history", txType: "Type", txAsset: "Related item", amount: "Amount", txHash: "Transaction", positive: "In", negative: "Out",
    convertEyebrow: "Asset conversion", convertTitle: "Convert an asset into release rights", convertLead: "Conversion grants distribution use in selected regions and channels without transferring ownership.",
    selectAsset: "Select asset", config: "License configuration", region: "Release scope", duration: "License term",
    regionSea: "Southeast Asia subtitles, dubbing and channel release", regionEurope: "Europe subtitles and streaming release", regionGlobal: "Global digital channels (non-exclusive)",
    day: "days", feeNotice: "Confirmation consumes 18 CLIP to create the regional license. Ownership remains unchanged.",
    acceptLicense: "I accept the licensing scope and commercial usage terms.", submitConversion: "Create release license",
    conversionDone: "License request submitted", unavailable: "Unavailable",
    transferEyebrow: "HAPW ASSET EXERCISE", transferTitle: "HAPW asset exercise", transferLead: "Choose a HAPW and a third-party platform to create a wallet signature request.",
    outbound: "Clipli → HAIWEN", inbound: "HAIWEN → Clipli", sourceAccount: "Source account", targetAccount: "Target account",
    transferDetails: "Exercise details", held: "Held", unit: "item", acceptOwnership: "I confirm my right to exercise this HAPW and authorize a wallet signature request.",
    transferNotice: "Final state follows the wallet, onchain record, and third-party rules.", submitTransfer: "Request exercise signature", transferDone: "Exercise request created",
    bindEyebrow: "HWF connect", bindTitle: "Bind HAIWEN account", bindLead: "Sync distribution level, channel benefits and services without sharing payment passwords or identity credentials.",
    overseasAccount: "HAIWEN account", authorizedAccount: "Authorized", phone: "Phone binding", points: "points", benefits: "Overseas distribution benefits active",
    priority: "Channel priority +1", discount: "Service fee -10%", quota: "Monthly localization +60 CLIP",
    bindAction: "Bind account", unbind: "Unbind", accountPlaceholder: "Enter HAIWEN account",
    walletEyebrow: "Web3 login", walletTitle: "Connect wallet", walletLead: "Choose a wallet for the demo login. The platform never asks for or stores private keys or seed phrases.",
    browserWallet: "Browser extension wallet", mobileWallet: "Mobile connection", walletService: "Web3 wallet service", selected: "Selected",
    connectSelected: "Connect selected wallet", connected: "Connected", chooseWallet: "Choose a wallet first",
    securityEyebrow: "Security center", securityTitle: "Security and authorization", securityLead: "Manage signature confirmation, license reminders and connected ecosystem accounts.",
    walletSign: "Wallet signature confirmation", walletSignDesc: "Asset conversions and transfers always require a wallet signature.", expiryReminder: "License expiry reminder",
    expiryReminderDesc: "Send an in-app reminder 7 days before expiry.", justNow: "just now", enabled: "Enabled", disabled: "Disabled",
    aboutEyebrow: "About Clipli", aboutTitle: "Put works, rights and release on one practical track", aboutLead: "Clipli gives creators, content holders and regional release teams a clear workbench for reviewing, licensing and operating digital work assets.",
    utilityTitle: "HAPW carries work rights", utilityText: "HAPW represents work assets issued on Clipli. It links a title, characters, rights metadata and release scope, so a holder can act on a specific work instead of an opaque number.",
    rightsTitle: "CLIP is the platform point", rightsText: "CLIP is issued by Clipli with no fixed cap. It supports license fees, creator rewards and service exchange. Users can purchase it through an external DEX, while every balance change remains visible in the dashboard.",
    platformRoles: "How the platform works", creatorRole: "For creators", creatorRoleDesc: "Publish a work, maintain rights information, set eligible regions and update notes when versions change.", distributorRole: "For release teams", distributorRoleDesc: "Run subtitles, dubbing, channel listing and local operations only inside a clear license.", holderRole: "For HAPW holders", holderRoleDesc: "Review asset status, request a license or transfer, and keep CLIP fees separate from work rights.", governanceTitle: "A grounded release rule", governanceText: "Clipli separates content, asset state and regional permission first, then reconnects them through explicit operation records. Future APIs, wallets and chain services must preserve the same fields and audit boundaries.",
    aboutWhyTitle: "Why Clipli exists", aboutWhyText1: "Moving a work into another market creates version, rights, subtitle, dubbing and channel questions at once. When those facts live across chat threads, spreadsheets and wallet addresses, nobody can quickly see who received which rights and when.", aboutWhyText2: "Clipli does something practical: it keeps work, rights and execution in one context, so people can see the boundary before continuing.",
    recordTitle: "What the platform records", recordIntro: "It does not add another label. It preserves facts that can still be checked later.", record1Title: "Work version", record1Desc: "Title, medium, creator, summary and change history.", record2Title: "Rights scope", record2Desc: "Linked HAPW, holder, eligible regions, channels and term.", record3Title: "Execution result", record3Desc: "CLIP fees, license requests, transfer state and platform entries.",
    boundaryTitle: "What Clipli does not do", boundaryText: "The platform does not custody private keys, guarantee CLIP prices, turn HAPW into automatic financial yield, or replace a formal copyright agreement. Demo data explains the product flow only.", principleTitle: "Working principles", principle1: "Confirm ownership before release", principle2: "Describe each region separately", principle3: "Keep fees separate from work assets", principle4: "Preserve traceable external entries",
    clipPageEyebrow: "CLIP MECHANISM · V1", clipPageTitle: "How CLIP moves through Clipli", clipPageLead: "CLIP is the unit used to price and settle platform services. It can be acquired on an external DEX or earned from creative and release tasks, but it never represents ownership of a work.", backAssets: "Back to asset dashboard", clipCurrentBalance: "Available balance", clipNoPeg: "No fixed USDT peg", clipContractPending: "Official contract details pending",
    clipMechanismTitle: "A simple usage loop", clipFlowBuy: "Swap from USDT", clipFlowBuyDesc: "Users acquire CLIP at a market price on an external DEX.", clipFlowUse: "Pay for services", clipFlowUseDesc: "Use it for release licensing, localization and platform services.", clipFlowReward: "Earn task rewards", clipFlowRewardDesc: "Clipli issues rewards for completed creative and release work.", clipFlowRecord: "Record the ledger", clipFlowRecordDesc: "Every inflow and outflow returns to the user's history.",
    clipRelationTitle: "What HAPW and CLIP each solve", hapwMechanismTitle: "HAPW: work rights", hapwMechanismText: "It maps to a work or rights bundle and records holder, licensable scope and transfer state.", clipMechanismRoleTitle: "CLIP: service settlement", clipMechanismRoleText: "It pays for platform services or carries task rewards, without ownership of any HAPW.", clipNoAutoTitle: "No automatic conversion", clipNoAutoText: "Creating HAPW does not mint CLIP, and holding CLIP does not grant work rights. The ledgers meet only when a business action such as a license fee requires credits and CLIP.",
    clipExchangeTitle: "Swap USDT for CLIP", clipExchangeLead: "The swap happens on an external DEX. Clipli provides an entry and instructions, but does not custody funds or promise a fixed rate.", exchange1: "Connect a wallet on the supported network", exchange2: "Choose USDT as the input asset", exchange3: "Verify the official CLIP contract and slippage", exchange4: "Confirm the swap and wait for settlement", marketPrice: "Liquidity sets the price", marketPriceDesc: "CLIP is not a stablecoin. The CLIP/USDT rate changes with available liquidity.",
    clipPolicyTitle: "Mechanism boundaries", supplyPolicyTitle: "Issuance", supplyPolicyText: "There is no fixed cap. Clipli releases CLIP only against real service demand, completed tasks and an incentive budget, while keeping issuance records.", usagePolicyTitle: "Utility", usagePolicyText: "Current uses include license fees, creator rewards, localization and platform service exchange.", recyclePolicyTitle: "Return flow", recyclePolicyText: "Service fees enter the platform service account and may fund operations or future rewards. There is no promised burn or fixed buyback.", disclosurePolicyTitle: "Disclosure", disclosurePolicyText: "Before launch, Clipli should publish the network, contract, liquidity pool and issuance record.", clipDemoNotice: "This is a functional demo. The official CLIP contract and network are not configured. Do not swap based on the token name alone.",
    clipNumbersTitle: "A few verifiable demo numbers", clipNumbersLead: "These values come from the current demo ledger and show where CLIP changes hands.", clipMetricBalance: "Current balance", clipMetricLicense: "Single license sample", clipMetricLocalization: "Localization sample", clipMetricReward: "Creator reward sample", clipMetricRecords: "Ledger records", clipMetricQuote: "External quote asset",
    clipPlayTitle: "Four common use cases", clipPlayLead: "Think of CLIP as an operating allowance for platform services, not a replacement for HAPW.", clipPlay1Title: "Regional license", clipPlay1Desc: "Select a HAPW, region and 30/90/180-day term; the current demo charges 18 CLIP per request.", clipPlay2Title: "Localization service", clipPlay2Desc: "Subtitles, dubbing and channel preparation can be billed separately; the demo shows a 60 CLIP sample.", clipPlay3Title: "Creator reward", clipPlay3Desc: "Completed creative or release tasks can earn CLIP; the demo shows a 240 CLIP sample.", clipPlay4Title: "Hold or swap", clipPlay4Desc: "Keep the balance for services or swap on an external DEX. Price follows liquidity and does not create HAPW ownership.",
    aboutWorkflowTitle: "From first registration to regional launch", aboutWorkflowLead: "One work can have several regional versions, but every version should start from the same rights note.", aboutWorkflow1Title: "Make the work legible", aboutWorkflow1Desc: "Capture title, creator, medium, version and rights notes as a readable work record.", aboutWorkflow2Title: "Separate the rights", aboutWorkflow2Desc: "Describe holdable HAPW, eligible regions and operating term as separate facts.", aboutWorkflow3Title: "Hand it to release", aboutWorkflow3Desc: "Subtitles, dubbing, channel listing and takedown return to the license record instead of oral agreements.",
    success: "Success", failed: "Something went wrong", noData: "No data", yuan: "¥"
  }
};

const extraLocales = window.CLIPLI_LOCALES || {};
Object.assign(translations.zh, extraLocales.zh || {});
Object.assign(translations.en, extraLocales.en || {});
translations.es = Object.assign({}, translations.en, extraLocales.es || {});
translations.ja = Object.assign({}, translations.en, extraLocales.ja || {});
translations.fr = Object.assign({}, translations.en, extraLocales.fr || {});
translations.ko = Object.assign({}, translations.en, extraLocales.ko || {});

const state = {
  lang: localStorage.getItem("clipli-lang") || "zh",
  theme: document.documentElement.dataset.theme || "dark",
  route: "/",
  menuOpen: false,
  worksFilter: "all",
  selectedAsset: "",
  conversionDays: 90,
  selectedPlatform: "haiwen",
  selectedWallet: "",
  generationDuration: 15,
  generationQuality: "standard",
  assetGuide: readAssetGuideProgress(),
  assetGuideFeedbackKey: ""
};

const navItems = [
  ["/", "navHome"],
  ["/works", "navWorks"],
  ["/assets", "navAssets"],
  ["/studio", "navStudio"],
  ["/transfer", "navTransfer"],
  ["/about", "navAbout"]
];

function t(key) {
  return (translations[state.lang] || translations.zh)[key] || translations.zh[key] || key;
}

function escapeHtml(value) {
  return String(value ?? "").replace(/[&<>"']/g, character => ({
    "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#039;"
  })[character]);
}

function safeExternalUrl(value) {
  try {
    const url = new URL(String(value));
    return ["http:", "https:"].includes(url.protocol) ? escapeHtml(url.href) : "#";
  } catch {
    return "#";
  }
}

function operationId(prefix) {
  const id = window.crypto && typeof window.crypto.randomUUID === "function"
    ? window.crypto.randomUUID()
    : Date.now() + "-" + Math.random().toString(16).slice(2);
  return prefix + "-" + id;
}

const ASSET_GUIDE_STORAGE_KEY = "clipli-asset-guide-v1";

function readAssetGuideProgress() {
  try {
    const saved = JSON.parse(localStorage.getItem(ASSET_GUIDE_STORAGE_KEY) || "{}");
    return {
      redeem: saved.redeem === true,
      exchange: saved.exchange === true,
      dex: saved.dex === true
    };
  } catch {
    return { redeem: false, exchange: false, dex: false };
  }
}

function updateAssetGuide(step, feedbackKey) {
  state.assetGuide = Object.assign({}, state.assetGuide, { [step]: true });
  state.assetGuideFeedbackKey = feedbackKey;
  localStorage.setItem(ASSET_GUIDE_STORAGE_KEY, JSON.stringify(state.assetGuide));
}

function resetAssetGuide() {
  state.assetGuide = { redeem: false, exchange: false, dex: false };
  state.assetGuideFeedbackKey = "guideFeedbackReset";
  localStorage.removeItem(ASSET_GUIDE_STORAGE_KEY);
}

function local(item, key) {
  if (!item) return "";
  if (state.lang === "zh") return escapeHtml(item[key]);
  const suffix = { en: "En", es: "Es", ja: "Ja", fr: "Fr", ko: "Ko" }[state.lang] || "En";
  return escapeHtml(item[key + suffix] || item[key + "En"] || item[key]);
}

function statusLabel(item) {
  const keys = { completed: "statusCompleted", pending: "statusPending", archived: "statusArchived", submitted: "statusSubmitted", awaitingSignature: "statusAwaitingSignature" };
  return item.statusCode && keys[item.statusCode] ? t(keys[item.statusCode]) : local(item, "status");
}

function transactionTypeLabel(item) {
  const keys = { playback: "txPlayback", reward: "txReward", license: "txLicense", dex: "txDex", localization: "txLocalization", redemptionGrant: "txRedemptionGrant", generationFee: "txGenerationFee", hapwExchange: "txHapwExchange" };
  return item.typeCode && keys[item.typeCode] ? t(keys[item.typeCode]) : local(item, "type");
}

function transferDirectionLabel(item) {
  return local(item, "direction");
}

function money(value) {
  const locale = { zh: "zh-CN", en: "en-US", es: "es-ES", ja: "ja-JP", fr: "fr-FR", ko: "ko-KR" }[state.lang] || "zh-CN";
  return new Intl.NumberFormat(locale, { maximumFractionDigits: 0 }).format(value);
}

function api(path, options) {
  return fetch(path, Object.assign({
    headers: { "Content-Type": "application/json" }
  }, options || {})).then(async response => {
    const payload = await response.json();
    if (!response.ok) {
      const code = payload.error && payload.error.code;
      const key = code ? "error_" + code : "failed";
      const message = t(key) === key ? t("failed") : t(key);
      const requestError = new Error(message);
      requestError.code = code;
      throw requestError;
    }
    return payload.data;
  });
}

function pageHead(eyebrow, title, lead, action) {
  return [
    '<header class="page-head"><div><p class="eyebrow">', eyebrow, '</p><h1>', title,
    '</h1><p class="lede">', lead, '</p></div>', action || "", '</header>'
  ].join("");
}

function header() {
  const nav = navItems.map(item => {
    const active = state.route === item[0]
      || (item[0] === "/works" && state.route.startsWith("/work/"))
      || (item[0] === "/assets" && ["/convert", "/bind", "/security", "/clip"].includes(state.route));
    return '<a class="nav-link ' + (active ? "active" : "") + '" href="#' + item[0] + '"' + (active ? ' aria-current="page"' : '') + '>' + t(item[1]) + "</a>";
  }).join("");
  return [
    '<header class="site-header"><div class="nav-shell">',
    '<a class="brand" href="#/" aria-label="Clipli · Create, click, grow"><span class="brand-mark">C</span><span class="brand-word"><b>Clipli</b><small>Create · Click · Grow</small></span></a>',
    '<nav class="nav-links ', state.menuOpen ? "open" : "", '" aria-label="', t("primaryNavigation"), '">', nav, '</nav>',
    '<div class="nav-actions"><label class="lang-switch"><span class="visually-hidden">', t("language"), '</span><select class="language-select" data-language aria-label="', t("language"), '">',
    [["zh", "中文"], ["en", "EN"], ["es", "ES"], ["ja", "日本語"], ["fr", "FR"], ["ko", "한국어"]].map(item => '<option value="' + item[0] + '" ' + (state.lang === item[0] ? "selected" : "") + '>' + item[1] + '</option>').join(""),
    '</select></label>',
    '<button class="icon-button theme-button" type="button" data-theme-toggle aria-label="', t(state.theme === "dark" ? "themeLight" : "themeDark"), '" title="', t(state.theme === "dark" ? "themeLight" : "themeDark"), '"><span aria-hidden="true">', state.theme === "dark" ? "☀" : "☾", '</span></button>',
    '<a class="icon-button" href="#/wallet" aria-label="', t("connect"), '" title="', t("connect"), '">◇</a>',
    '<button class="icon-button menu-button" id="menu-toggle" type="button" aria-label="', t("menu"), '" aria-expanded="', state.menuOpen, '">≡</button>',
    '</div></div></header>'
  ].join("");
}

function footer() {
  return '<footer class="site-footer"><span>© 2026 Clipli</span><span>' + t("footer") + "</span></footer>";
}

function shell(content) {
  app.innerHTML = header() + '<main id="main" class="main">' + content + "</main>" + footer();
  const skipLink = document.querySelector(".skip-link");
  if (skipLink) skipLink.textContent = t("skipToContent");
  bindGlobal();
}

function bindGlobal() {
  document.querySelectorAll("[data-language]").forEach(select => select.addEventListener("change", () => {
    state.lang = select.value;
    localStorage.setItem("clipli-lang", state.lang);
    render();
  }));
  const themeToggle = document.querySelector("[data-theme-toggle]");
  if (themeToggle) themeToggle.addEventListener("click", () => {
    state.theme = state.theme === "dark" ? "light" : "dark";
    document.documentElement.dataset.theme = state.theme;
    localStorage.setItem("clipli-theme", state.theme);
    const themeColor = document.querySelector('meta[name="theme-color"]');
    if (themeColor) themeColor.content = state.theme === "dark" ? "#11110f" : "#f2efe7";
    render();
  });
  const menu = document.getElementById("menu-toggle");
  if (menu) menu.addEventListener("click", () => {
    state.menuOpen = !state.menuOpen;
    document.querySelector(".nav-links").classList.toggle("open", state.menuOpen);
    menu.setAttribute("aria-expanded", String(state.menuOpen));
  });
  document.querySelectorAll(".nav-link").forEach(link => link.addEventListener("click", () => { state.menuOpen = false; }));
}

function toast(message, type) {
  const node = document.createElement("div");
  node.className = "toast " + (type || "");
  node.textContent = message;
  toastRegion.appendChild(node);
  setTimeout(() => node.remove(), 3200);
}

function loading() {
  shell('<div class="loading">' + t("loading") + "</div>");
}

async function renderHome() {
  const [data, works, assetData] = await Promise.all([api("/api/overview"), api("/api/works"), api("/api/assets")]);
  const metrics = [
    [money(data.counts.authorized), t("authorized")],
    [money(data.counts.overseas), t("overseas")],
    [data.counts.visible + "%", t("visibility")]
  ].map(item => '<div class="metric"><strong>' + item[0] + '</strong><span>' + item[1] + "</span></div>").join("");
  const featureKeys = [["01", "feature1", "feature1Desc"], ["02", "feature2", "feature2Desc"], ["03", "feature3", "feature3Desc"]];
  const featureIcons = ["◈", "⌁", "◎"];
  const features = featureKeys.map((item, index) => '<article class="feature"><span class="feature-icon" aria-hidden="true">' + featureIcons[index] + '</span><span class="feature-index">' + item[0] + '</span><h3>' + t(item[1]) + '</h3><p>' + t(item[2]) + "</p></article>").join("");
  const steps = [["01", "step1", "step1Desc"], ["02", "step2", "step2Desc"], ["03", "step3", "step3Desc"], ["04", "step4", "step4Desc"]].map(item => '<article class="step"><b>' + item[0] + '</b><h3>' + t(item[1]) + '</h3><p>' + t(item[2]) + "</p></article>").join("");
  const signals = ["signal1", "signal2", "signal3", "signal4"].map((key, index) => '<span><b>0' + (index + 1) + '</b>' + t(key) + '</span>').join("");
  const workRail = works.slice(0, 4).map((work, index) => [
    '<a class="rail-work" data-accent="', escapeHtml(work.accent), '" href="#/work/', encodeURIComponent(work.id), '"><div class="rail-art"><span>0', index + 1,
    '</span><i></i></div><div class="rail-copy"><small>', local(work, "category"), ' · ', escapeHtml(work.duration), '</small><h3>', local(work, "title"),
    '</h3><p>', local(work, "summary"), '</p><strong>', t("details"), ' →</strong></div></a>'
  ].join("")).join("");
  const practices = [["01", "practice1", "practice1Desc"], ["02", "practice2", "practice2Desc"], ["03", "practice3", "practice3Desc"]].map(item => [
    '<div class="practice-row"><b>', item[0], '</b><div><h3>', t(item[1]), '</h3><p>', t(item[2]), '</p></div></div>'
  ].join("")).join("");
  const activityBars = [[t("activityWorks"), works.length, Math.min(100, works.length * 18)], [t("activityAssets"), assetData.assets.filter(item => item.redemptionStatus === "available").length, Math.min(100, assetData.stats.holdings * 12)], [t("activityLedger"), assetData.generations.length, Math.min(100, assetData.generations.length * 32)]]
    .map(item => ['<div class="activity-row"><div><span>', item[0], '</span><b>', item[1], '</b></div><i style="--bar-width:', item[2], '%"></i></div>'].join("")).join("");
  const asset = data.featuredAsset;
  const brandPillars = [["Create", "brandCreate"], ["Click", "brandClick"], ["Grow", "brandGrow"]]
    .map(item => '<span><b>' + item[0] + '</b><small>' + t(item[1]) + '</small></span>').join("");
  shell([
    '<section class="hero"><div class="hero-copy"><p class="eyebrow">', t("homeEyebrow"), '</p><h1><span>Clipli</span><span>', t("homeTitle"), '</span></h1><p class="lede">', t("homeLead"),
    '</p><div class="brand-pillars">', brandPillars, '</div><div class="hero-actions"><a class="button primary" href="#/studio">', t("openStudio"), '<span class="button-icon">→</span></a><a class="button" href="#/works">', t("browseWorks"), '</a></div>',
    '<div class="hero-metrics">', metrics, '</div></div>',
    '<div class="iso-stage" aria-label="Isometric Clipli asset"><div class="iso-grid"></div><div class="orbit orbit-one"></div><div class="orbit orbit-two"></div><div class="iso-tower"><div class="cube one"><span></span></div><div class="cube two"><span></span></div><div class="cube three"><span></span></div></div>',
    '<a class="asset-float" href="#/studio"><small>', t("creditYield"), '</small><strong>HAPW ', escapeHtml(asset.tokenId), ' · ', asset.creditYield, ' ', t("creditsUnit"), '</strong></a></div></section>',
    '<section class="signal-strip" aria-label="', t("deskStatus"), '"><strong>', t("deskStatus"), '</strong><div class="signal-window"><div class="signal-track"><div class="signal-set">', signals, '</div><div class="signal-set" aria-hidden="true">', signals, '</div></div></div></section>',
    '<section class="system-flow five" aria-label="Clipli workflow"><a class="flow-node" href="#/works"><span class="flow-icon">W</span><b>HAPW</b><small>', t("studioFlow1"), '</small></a><span class="flow-arrow">→</span><a class="flow-node" href="#/studio"><span class="flow-icon">✓</span><b>', t("studioFlow2"), '</b><small>', t("feature2Desc"), '</small></a><span class="flow-arrow">→</span><a class="flow-node" href="#/studio"><span class="flow-icon">▶</span><b>', t("studioFlow3"), '</b><small>', t("generationCreditsLead"), '</small></a><span class="flow-arrow">→</span><a class="flow-node" href="#/clip"><span class="flow-icon">◎</span><b>', t("studioFlow4"), '</b><small>', t("feature3Desc"), '</small></a><span class="flow-arrow">→</span><a class="flow-node" href="#/assets"><span class="flow-icon">P</span><b>CLIP</b><small>', t("studioFlow5"), '</small></a></section>',
    '<section class="home-insights"><div class="insight-chart"><div class="insight-head"><div><p class="eyebrow">ACTIVITY PULSE</p><h2>', t("pulseTitle"), '</h2><p>', t("pulseDesc"), '</p></div><strong>', data.counts.visible, '%</strong></div><svg class="home-chart" viewBox="0 0 420 150" role="img" aria-label="', t("pulseDesc"), '"><g class="chart-grid"><line x1="0" y1="30" x2="420" y2="30"></line><line x1="0" y1="75" x2="420" y2="75"></line><line x1="0" y1="120" x2="420" y2="120"></line></g><polyline points="0,112 70,96 140,103 210,64 280,75 350,37 420,28"></polyline><circle cx="420" cy="28" r="5"></circle></svg><div class="chart-legend"><span><i></i>', t("pulseAuthorized"), '</span><span><i></i>', t("pulseOverseas"), '</span><span><i></i>', t("pulseVisibility"), '</span></div></div><div class="insight-activity"><p class="eyebrow">LEDGER CHECK</p><h2>', t("activityTitle"), '</h2><p>', t("activityDesc"), '</p><div class="activity-list">', activityBars, '</div></div></section>',
    '<section class="home-works"><header class="home-section-head"><div><p class="eyebrow">CURRENT SLATE</p><h2>', t("selectedWorks"), '</h2><p>', t("selectedLead"), '</p></div><a class="text-link" href="#/works">', t("viewAllWorks"), ' →</a></header><div class="work-rail">', workRail, '</div></section>',
    '<section class="home-practice"><div class="practice-intro"><p class="eyebrow">', t("practiceEyebrow"), '</p><h2>', t("practiceTitle"), '</h2><p>', t("practiceLead"), '</p><div class="practice-links"><a class="text-link" href="#/studio">', t("openStudio"), ' →</a><a class="text-link" href="#/clip">', t("homeClipLink"), ' →</a></div></div><div class="practice-list">', practices, '</div></section>',
    '<section class="feature-band"><div class="feature-grid">', features, '</div></section>',
    '<section class="home-final"><div class="home-boundary"><div class="boundary-copy"><p class="eyebrow">OPERATING BOUNDARY</p><h2>', t("homeBoundaryTitle"), '</h2></div><p class="boundary-text">', t("homeBoundaryText"), '</p></div><div><h2 class="section-title">', t("releaseTitle"), '</h2><div class="steps">', steps, "</div></div></section>"
  ].join(""));
}

async function renderWorks() {
  const works = await api("/api/works");
  const filtered = state.worksFilter === "featured" ? works.slice(0, 3) : works;
  const cards = filtered.map(work => [
    '<article class="work-card" data-accent="', escapeHtml(work.accent), '" data-detail="', escapeHtml(work.id), '" tabindex="0" role="link" aria-label="', t("details"), ' ', local(work, "title"), '"><div class="work-visual"><span class="work-duration">', escapeHtml(work.duration),
    '</span><button class="play-button" type="button" data-play="', escapeHtml(work.id), '" aria-label="', t("play"), ' ', local(work, "title"), '">▶</button></div>',
    '<div class="work-body"><span class="tag">', local(work, "category"), '</span><h2>', local(work, "title"), '</h2><p>', local(work, "summary"), '</p>',
    '<div class="work-meta"><span>', local(work, "creator"), '</span><span>', money(work.views), " ", t("views"), '</span></div><a class="text-link work-detail-link" href="#/work/', encodeURIComponent(work.id), '">', t("details"), ' →</a></div></article>'
  ].join("")).join("");
  const controls = '<div class="filter-row"><button class="filter-button ' + (state.worksFilter === "all" ? "active" : "") + '" data-filter="all">' + t("all") + '</button><button class="filter-button ' + (state.worksFilter === "featured" ? "active" : "") + '" data-filter="featured">' + t("featured") + "</button></div>";
  shell(pageHead(t("worksEyebrow"), t("worksTitle"), t("worksLead"), controls) + '<section class="work-grid">' + cards + "</section>");
  document.querySelectorAll("[data-filter]").forEach(button => button.addEventListener("click", () => { state.worksFilter = button.dataset.filter; renderWorks(); }));
  document.querySelectorAll("[data-detail]").forEach(card => {
    const openDetail = event => {
      if (event.target.closest("button, a")) return;
      location.hash = "#/work/" + card.dataset.detail;
    };
    card.addEventListener("click", openDetail);
    card.addEventListener("keydown", event => {
      if (event.key === "Enter" || event.key === " ") {
        event.preventDefault();
        location.hash = "#/work/" + card.dataset.detail;
      }
    });
  });
  document.querySelectorAll("[data-play]").forEach(button => button.addEventListener("click", () => {
    const work = works.find(item => item.id === button.dataset.play);
    modalRoot.innerHTML = [
      '<div class="modal-backdrop" role="presentation"><section class="modal" role="dialog" aria-modal="true" aria-labelledby="modal-title">',
      '<div class="modal-visual"><div class="modal-cube"></div></div><div class="modal-body"><div><h2 id="modal-title">', local(work, "title"),
      '</h2><p>', local(work, "summary"), ' · ', t("seedSource"), '</p></div><button class="button" id="modal-close">', t("close"), "</button></div></section></div>"
    ].join("");
    document.getElementById("modal-close").addEventListener("click", () => { modalRoot.innerHTML = ""; });
    modalRoot.querySelector(".modal-backdrop").addEventListener("click", event => { if (event.target.classList.contains("modal-backdrop")) modalRoot.innerHTML = ""; });
  }));
}

async function renderWorkDetail() {
  const workId = state.route.split("/")[2];
  const work = await api("/api/works/" + encodeURIComponent(workId));
  const assets = await api("/api/assets");
  const linked = assets.assets.find(asset => asset.id === work.linkedAssetId);
  shell([
    '<div class="detail-shell"><a class="back-link" href="#/works">← ', t("backWorks"), '</a><section class="detail-hero"><div class="detail-visual" data-accent="', escapeHtml(work.accent), '"><div class="detail-shape"></div><span class="work-duration">', escapeHtml(work.duration), '</span><button class="play-button" id="detail-play" type="button" aria-label="', t("play"), '">▶</button></div><div class="detail-copy"><p class="eyebrow">', t("detailEyebrow"), '</p><h1>', local(work, "title"), '</h1><p class="lede">', local(work, "summary"), '</p><div class="detail-meta"><span class="tag">', local(work, "category"), '</span><span>', local(work, "creator"), '</span><span>', money(work.views), ' ', t("views"), '</span></div></div></section><section class="detail-grid"><article class="detail-panel"><h2>', t("details"), '</h2><p>', local(work, "description"), '</p><div class="detail-facts"><div><small>', t("format"), '</small><strong>', local(work, "format"), '</strong></div><div><small>', t("linkedHapw"), '</small><strong>', linked ? 'HAPW ' + escapeHtml(linked.tokenId) + ' · ' + local(linked, "name") : t("noData"), '</strong></div><div><small>', t("thirdPartyRelease"), '</small><strong>', t("platformChoice"), '</strong></div></div></article><aside class="detail-panel access-panel"><p class="eyebrow">HAPW ACCESS</p><h2>', t("materialAccess"), '</h2><p>', t("materialAccessDesc"), '</p><dl class="access-facts"><div><dt>', t("rightsHolder"), '</dt><dd>', linked ? local(linked, "rightsHolder") : t("noData"), '</dd></div><div><dt>', t("authorizationScope"), '</dt><dd>', linked ? local(linked, "authorizationScope") : t("noData"), '</dd></div><div><dt>', t("creditYield"), '</dt><dd>', linked ? linked.creditYield + ' ' + t("creditsUnit") : t("noData"), '</dd></div></dl><a class="button secondary wide" href="#/studio">', t("openStudioWithWork"), ' →</a></aside><aside class="detail-panel release-panel"><p class="eyebrow">THIRD-PARTY PLATFORM</p><h2>', t("thirdPartyRelease"), '</h2><p>', t("thirdPartyLead"), '</p><button class="button primary wide" id="choose-platform" type="button">', t("choosePlatform"), ' ↗</button></aside></section></div>'
  ].join(""));
  document.getElementById("detail-play").addEventListener("click", () => openWorkModal(work));
  document.getElementById("choose-platform").addEventListener("click", () => openPlatformPicker(assets.platforms));
}

function openWorkModal(work) {
  modalRoot.innerHTML = [
    '<div class="modal-backdrop" role="presentation"><section class="modal" role="dialog" aria-modal="true" aria-labelledby="modal-title">',
    '<div class="modal-visual"><div class="modal-cube"></div></div><div class="modal-body"><div><h2 id="modal-title">', local(work, "title"), '</h2><p>', local(work, "summary"), ' · ', t("seedSource"), '</p></div><button class="button" id="modal-close">', t("close"), "</button></div></section></div>"
  ].join("");
  document.getElementById("modal-close").addEventListener("click", () => { modalRoot.innerHTML = ""; });
  modalRoot.querySelector(".modal-backdrop").addEventListener("click", event => { if (event.target.classList.contains("modal-backdrop")) modalRoot.innerHTML = ""; });
}

function openPlatformPicker(platforms) {
  const rows = platforms.map(platform => [
    '<a class="platform-option" href="', safeExternalUrl(platform.url), '" target="_blank" rel="noopener noreferrer">',
    '<span class="platform-monogram">', escapeHtml(platform.name.slice(0, 1)), '</span><span><strong>', escapeHtml(platform.name), '</strong><small>', t("platform_" + platform.code), '</small></span><b>↗</b></a>'
  ].join("")).join("");
  modalRoot.innerHTML = [
    '<div class="modal-backdrop"><section class="modal platform-modal" role="dialog" aria-modal="true" aria-labelledby="platform-title"><div class="platform-head"><div><p class="eyebrow">EXTERNAL DESTINATIONS</p><h2 id="platform-title">', t("choosePlatformTitle"), '</h2><p>', t("platformNotice"), '</p></div><button class="icon-button" id="platform-close" type="button" aria-label="', t("close"), '">×</button></div><div class="platform-list">', rows, '</div></section></div>'
  ].join("");
  const close = () => { modalRoot.innerHTML = ""; };
  document.getElementById("platform-close").addEventListener("click", close);
  modalRoot.querySelector(".modal-backdrop").addEventListener("click", event => { if (event.target.classList.contains("modal-backdrop")) close(); });
}

function openRedemptionConfirm(asset) {
  modalRoot.innerHTML = [
    '<div class="modal-backdrop"><section class="modal confirm-modal" role="dialog" aria-modal="true" aria-labelledby="redeem-confirm-title" aria-describedby="redeem-confirm-text">',
    '<div class="confirm-mark" aria-hidden="true">W</div><div class="confirm-content"><p class="eyebrow">HAPW REDEMPTION</p><h2 id="redeem-confirm-title">', t("redeemConfirmTitle"), '</h2><p id="redeem-confirm-text">', t("redeemConfirmText"), '</p>',
    '<dl class="confirm-summary"><div><dt>HAPW</dt><dd>', escapeHtml(asset.tokenId), ' · ', local(asset, "name"), '</dd></div><div><dt>', t("rightsHolder"), '</dt><dd>', local(asset, "rightsHolder"), '</dd></div><div><dt>', t("redemptionGrant"), '</dt><dd>', asset.creditYield, ' ', t("creditsUnit"), ' + ', Math.floor(asset.creditYield * .4), ' CLIP</dd></div></dl>',
    '<div class="confirm-actions"><button class="button" id="redeem-cancel" type="button">', t("cancel"), '</button><button class="button primary" id="redeem-confirm" type="button">', t("confirmRedeem"), '</button></div></div></section></div>'
  ].join("");
  const close = () => { modalRoot.innerHTML = ""; };
  document.getElementById("redeem-cancel").addEventListener("click", close);
  modalRoot.querySelector(".modal-backdrop").addEventListener("click", event => { if (event.target.classList.contains("modal-backdrop")) close(); });
  const confirm = document.getElementById("redeem-confirm");
  confirm.focus();
  confirm.addEventListener("click", async () => {
    confirm.disabled = true;
    try {
      await api("/api/hapw/redemptions", { method: "POST", body: JSON.stringify({ assetId: asset.id, accepted: true, requestId: operationId("redeem") }) });
      close();
      updateAssetGuide("redeem", "guideFeedbackRedeem");
      toast(t("redeemDone"));
      if (state.route === "/assets") renderAssets();
      else renderStudio();
    } catch (err) { toast(err.message, "error"); confirm.disabled = false; }
  });
}

async function renderAssets() {
  const data = await api("/api/assets");
  const pool = data.clip.dexPool;
  const exchangePolicy = data.clip.hapwExchangePolicy || { dailyLimit: 2, usedToday: 0, remainingToday: 2, inventoryTotal: 0, inventoryAvailable: 0, reached: false };
  const dailyLimitReached = exchangePolicy.reached || exchangePolicy.remainingToday <= 0;
  const stats = [
    [t("holdings"), data.stats.holdings, t("totalValue") + " " + t("yuan") + money(data.stats.totalValue)],
    [t("redeemedCount"), data.hapwRedemptions.length, t("creditsGranted") + " " + money(data.generationAccount.lifetimeGranted)],
    [t("generationJobs"), data.generations.length, t("creditsUsed") + " " + money(data.generationAccount.lifetimeUsed)],
    [t("clipSpent"), data.generationAccount.lifetimeClipSpent, "CLIP"]
  ].map(item => '<article class="stat-card"><span>' + item[0] + '</span><strong>' + String(item[1]).padStart(2, "0") + '</strong><small>' + item[2] + "</small></article>").join("");
  const rows = data.transfers.map(transfer => {
    const asset = data.assets.find(item => item.id === transfer.assetId);
    const pending = ["pending", "awaitingSignature"].includes(transfer.statusCode);
    return '<tr><td>HAPW ' + (asset ? escapeHtml(asset.tokenId) : "") + " · " + (asset ? local(asset, "name") : "") + '</td><td>' + transferDirectionLabel(transfer) + '</td><td>¥ ' + money(transfer.value) + '</td><td><span class="status ' + (pending ? "pending" : "") + '">' + statusLabel(transfer) + "</span></td><td>" + escapeHtml(transfer.createdAt) + "</td></tr>";
  }).join("");
  const clipRows = data.clipTransactions.map(transaction => '<tr><td>' + transactionTypeLabel(transaction) + '</td><td>' + local(transaction, "counterparty") + '</td><td class="amount ' + (transaction.amount >= 0 ? "in" : "out") + '">' + (transaction.amount >= 0 ? "+" : "") + money(transaction.amount) + ' CLIP</td><td>' + escapeHtml(transaction.txHash) + '</td><td>' + escapeHtml(transaction.createdAt) + '</td></tr>').join("");
  const generationRows = data.generations.map(item => {
    const asset = data.assets.find(assetItem => assetItem.id === item.assetId);
    return '<tr><td>' + local(item, "title") + '</td><td>HAPW ' + (asset ? escapeHtml(asset.tokenId) : "") + '</td><td>' + item.creditsUsed + ' ' + t("creditsUnit") + '</td><td class="amount out">-' + item.clipCost + ' CLIP</td><td>' + money(item.validViews) + '</td><td>' + statusLabel(item) + '</td><td>' + escapeHtml(item.createdAt) + '</td></tr>';
  }).join("");
  const reserveRows = data.assets.filter(item => item.clipPrice > 0).map(asset => {
    const fee = Math.ceil(asset.clipPrice * data.clip.hapwExchangeFeeRate);
    const unavailableLabel = asset.exchangeAvailable ? (dailyLimitReached ? t("reserveDailyLimit") : t("exchangeToHapw")) : t("reserveClaimed");
    return '<article class="reserve-row"><div><small>HAPW ' + escapeHtml(asset.tokenId) + '</small><h3>' + local(asset, "name") + '</h3><p>' + local(asset, "authorizationScope") + '</p></div><dl><div><dt>' + t("reservePrice") + '</dt><dd>' + money(asset.clipPrice) + ' CLIP</dd></div><div><dt>' + t("exchangeFee") + '</dt><dd>' + money(fee) + ' CLIP · 5%</dd></div></dl><button class="button secondary" type="button" data-hapw-exchange="' + escapeHtml(asset.id) + '" ' + (asset.exchangeAvailable && !dailyLimitReached ? "" : "disabled") + '>' + unavailableLabel + '</button></article>';
  }).join("");
  const redeemableAsset = data.assets.find(asset => asset.redemptionStatus === "available");
  const reserveAsset = data.assets
    .filter(asset => asset.exchangeAvailable && asset.clipPrice > 0)
    .sort((first, second) => first.clipPrice - second.clipPrice)[0];
  const reserveTotal = reserveAsset
    ? reserveAsset.clipPrice + Math.ceil(reserveAsset.clipPrice * data.clip.hapwExchangeFeeRate)
    : 0;
  const guideDoneCount = Object.values(state.assetGuide).filter(Boolean).length;
  const guideFeedbackKey = state.assetGuideFeedbackKey
    || (guideDoneCount === 3 ? "guideFeedbackAll" : guideDoneCount > 0 ? "guideFeedbackContinue" : "guideFeedbackIdle");
  const guideSteps = [
    '<article class="asset-guide-step ' + (state.assetGuide.redeem ? "is-done" : "") + '"><div class="guide-step-meta"><span>01</span><b>' + t(state.assetGuide.redeem ? "guideStatusDone" : "guideStatusReady") + '</b></div><div class="guide-route" aria-hidden="true"><span>W</span><i>→</i><span>P</span></div><h3>' + t("guide1Title") + '</h3><p>' + t("guide1Desc") + '</p><div class="guide-step-action"><small>' + t("guide1Quote") + '</small><strong>' + (redeemableAsset ? redeemableAsset.creditYield + ' ' + t("creditsUnit") + ' + ' + Math.floor(redeemableAsset.creditYield * data.generationAccount.clipGrantPerCredit) + ' CLIP' : t("guide1Unavailable")) + '</strong><button class="button primary" type="button" data-guide-redeem ' + (redeemableAsset ? "" : "disabled") + '><span class="button-icon">✓</span>' + t("guide1Action") + '</button></div></article>',
    '<article class="asset-guide-step ' + (state.assetGuide.exchange ? "is-done" : "") + '"><div class="guide-step-meta"><span>02</span><b>' + t(state.assetGuide.exchange ? "guideStatusDone" : "guideStatusReady") + '</b></div><div class="guide-route" aria-hidden="true"><span>P</span><i>→</i><span>W</span></div><h3>' + t("guide2Title") + '</h3><p>' + t("guide2Desc") + '</p><div class="guide-step-action"><small>' + t("guide2Quote") + '</small><strong>' + (dailyLimitReached ? t("guide2DailyLimit") : reserveAsset ? money(reserveTotal) + ' CLIP · ' + escapeHtml(reserveAsset.tokenId) : t("guide2Unavailable")) + '</strong><button class="button secondary" type="button" data-guide-exchange ' + (reserveAsset && !dailyLimitReached ? "" : "disabled") + '><span class="button-icon">⇄</span>' + (dailyLimitReached ? t("reserveDailyLimit") : t("guide2Action")) + '</button></div></article>',
    '<article class="asset-guide-step ' + (state.assetGuide.dex ? "is-done" : "") + '"><div class="guide-step-meta"><span>03</span><b>' + t(state.assetGuide.dex ? "guideStatusHandoff" : "guideStatusReady") + '</b></div><div class="guide-route" aria-hidden="true"><span>P</span><i>→</i><span>$</span></div><h3>' + t("guide3Title") + '</h3><p>' + t("guide3Desc") + '</p><div class="guide-step-action"><small>' + t("guide3Quote") + '</small><strong>1 CLIP ≈ ' + pool.usdtPerClip.toFixed(3) + ' USDT</strong><button class="button" type="button" data-dex-handoff><span class="button-icon">↗</span>' + t("guide3Action") + '</button></div></article>'
  ].join("");
  const guide = '<section class="asset-guide" aria-labelledby="asset-guide-title"><header><div><p class="eyebrow">FIRST RUN</p><h2 id="asset-guide-title">' + t("assetGuideTitle") + '</h2><p>' + t("assetGuideLead") + '</p></div><div class="guide-progress"><strong>' + guideDoneCount + '/3</strong><span>' + t("assetGuideProgress") + '</span><button class="text-link text-button" type="button" data-guide-reset><span aria-hidden="true">↻</span>' + t("assetGuideReset") + '</button></div></header><div class="asset-guide-track">' + guideSteps + '</div><footer class="guide-feedback" aria-live="polite"><span aria-hidden="true">◆</span><p>' + t(guideFeedbackKey) + '</p></footer></section>';
  const actions = '<div class="filter-row"><a class="button primary" href="#/studio">' + t("openStudio") + '</a><a class="button" href="#/transfer">' + t("transferAsset") + "</a></div>";
  shell([
    pageHead(t("assetsEyebrow"), t("assetsTitle"), t("assetsLead"), actions),
    guide,
    '<section class="ledger-grid three"><article class="token-card hapw-card"><div class="token-heading"><span class="token-mark">W</span><div><p class="eyebrow">HAPW</p><h2>', t("hapwHoldings"), '</h2></div></div><p>', t("hapwLead"), '</p><div class="token-number">', data.stats.holdings, '<small> ', t("holdings"), '</small></div><div class="token-foot"><span>', t("redeemedCount"), '</span><strong>', data.hapwRedemptions.length, '</strong></div></article>',
    '<article class="token-card credit-card"><div class="token-heading"><span class="token-mark">C</span><div><p class="eyebrow">CREATION CREDITS</p><h2>', t("generationCredits"), '</h2></div></div><p>', t("generationCreditsLead"), '</p><div class="token-number">', money(data.generationAccount.balance), '<small> ', t("creditsUnit"), '</small></div><div class="token-foot"><span>', t("creditsUsed"), '</span><strong>', money(data.generationAccount.lifetimeUsed), '</strong></div></article>',
    '<article class="token-card clip-card"><div class="token-heading"><span class="token-mark">P</span><div><p class="eyebrow">CLIP</p><h2>', t("clipBalance"), '</h2></div></div><p>', t("clipLead"), '</p><div class="token-number">', money(data.clip.balance), '<small> CLIP</small></div><div class="dex-mini"><span>', t("indicativeRate"), '</span><strong>1 USDT ≈ ', pool.clipPerUsdt.toFixed(2), ' CLIP</strong><small>', t("poolLiquidity"), ' · ', money(pool.totalLiquidityUsdt), ' USDT</small></div><div class="token-foot"><a class="text-link" href="#/clip">', t("mechanismDetails"), ' →</a><button class="button secondary" type="button" data-dex-handoff>', t("dexButton"), ' ↗</button></div></article></section>',
    '<section class="stat-grid four">', stats, '</section>',
    '<section class="reserve-section" id="hapw-reserve"><header><div><p class="eyebrow">CLIP → HAPW</p><h2>', t("reserveTitle"), '</h2><p>', t("reserveLead"), '</p></div><div class="reserve-policy"><span class="fee-badge">5% ', t("exchangeFee"), '</span><span class="daily-limit"><strong>', exchangePolicy.remainingToday, '/', exchangePolicy.dailyLimit, '</strong><small>', t("dailyRemaining"), '</small></span><span class="inventory-limit"><strong>', exchangePolicy.inventoryAvailable, '/', exchangePolicy.inventoryTotal, '</strong><small>', t("reserveAvailable"), '</small></span></div></header><div class="reserve-list">', reserveRows, '</div></section>',
    '<section class="data-section"><div class="data-head"><h2>', t("generationHistory"), '</h2><a class="text-link" href="#/studio">', t("openStudio"), ' →</a></div><div class="table-wrap"><table><thead><tr><th>', t("video"), '</th><th>HAPW</th><th>', t("creditsCost"), '</th><th>', t("clipCost"), '</th><th>', t("observedViews"), '</th><th>', t("status"), '</th><th>', t("date"), '</th></tr></thead><tbody>', generationRows, '</tbody></table></div></section>',
    '<section class="data-section"><div class="data-head"><h2>', t("recentTransfers"), '</h2><a class="text-link" href="#/transfer">', t("transferAsset"), ' →</a></div><div class="table-wrap"><table><thead><tr><th>', t("asset"), '</th><th>', t("exerciseTarget"), '</th><th>', t("value"), '</th><th>', t("status"), '</th><th>', t("date"), '</th></tr></thead><tbody>', rows, '</tbody></table></div></section>',
    '<section class="data-section"><div class="data-head"><div><h2>', t("clipHistory"), '</h2><p class="table-note">', t("clipAcquisitionSummary"), '</p></div><button class="text-link text-button" type="button" data-dex-handoff>', t("dexButton"), ' ↗</button></div><div class="table-wrap"><table><thead><tr><th>', t("txType"), '</th><th>', t("txAsset"), '</th><th>', t("amount"), '</th><th>', t("txHash"), '</th><th>', t("date"), '</th></tr></thead><tbody>', clipRows, '</tbody></table></div></section>'
  ].join(""));
  const guideRedeem = document.querySelector("[data-guide-redeem]");
  if (guideRedeem && redeemableAsset) guideRedeem.addEventListener("click", () => openRedemptionConfirm(redeemableAsset));
  const guideExchange = document.querySelector("[data-guide-exchange]");
  if (guideExchange && reserveAsset && !dailyLimitReached) guideExchange.addEventListener("click", () => openHapwExchangeConfirm(reserveAsset, data.clip.hapwExchangeFeeRate, exchangePolicy));
  document.querySelectorAll("[data-dex-handoff]").forEach(button => button.addEventListener("click", () => openDexHandoff(data.clip)));
  const guideReset = document.querySelector("[data-guide-reset]");
  if (guideReset) guideReset.addEventListener("click", () => { resetAssetGuide(); renderAssets(); });
  document.querySelectorAll("[data-hapw-exchange]").forEach(button => button.addEventListener("click", () => {
    const asset = data.assets.find(item => item.id === button.dataset.hapwExchange);
    if (asset && !dailyLimitReached) openHapwExchangeConfirm(asset, data.clip.hapwExchangeFeeRate, exchangePolicy);
  }));
}

function openHapwExchangeConfirm(asset, feeRate, exchangePolicy) {
  const fee = Math.ceil(asset.clipPrice * feeRate);
  const total = asset.clipPrice + fee;
  const policy = exchangePolicy || { dailyLimit: 2, usedToday: 0, remainingToday: 2 };
  modalRoot.innerHTML = [
    '<div class="modal-backdrop"><section class="modal confirm-modal" role="dialog" aria-modal="true" aria-labelledby="hapw-exchange-title"><div class="confirm-mark">P</div><div class="confirm-content"><p class="eyebrow">CLIP → HAPW</p><h2 id="hapw-exchange-title">', t("exchangeConfirmTitle"), '</h2><p>', t("exchangeConfirmText"), '</p><dl class="confirm-summary"><div><dt>HAPW</dt><dd>', escapeHtml(asset.tokenId), ' · ', local(asset, "name"), '</dd></div><div><dt>', t("reservePrice"), '</dt><dd>', asset.clipPrice, ' CLIP</dd></div><div><dt>', t("exchangeFee"), '</dt><dd>', fee, ' CLIP · 5%</dd></div><div><dt>', t("totalCost"), '</dt><dd>', total, ' CLIP</dd></div><div><dt>', t("dailyLimit"), '</dt><dd>', policy.usedToday, ' / ', policy.dailyLimit, ' · ', t("dailyRemaining"), ' ', policy.remainingToday, '</dd></div></dl><label class="check"><input id="hapw-exchange-accepted" type="checkbox"><span>', t("acceptHapwExchange"), '</span></label><div class="confirm-actions"><button class="button" id="hapw-exchange-cancel" type="button">', t("cancel"), '</button><button class="button primary" id="hapw-exchange-confirm" type="button">', t("confirmExchange"), '</button></div></div></section></div>'
  ].join("");
  const close = () => { modalRoot.innerHTML = ""; };
  document.getElementById("hapw-exchange-cancel").addEventListener("click", close);
  document.getElementById("hapw-exchange-confirm").addEventListener("click", async event => {
    const button = event.currentTarget;
    button.disabled = true;
    try {
      await api("/api/clip/hapw-exchanges", { method: "POST", body: JSON.stringify({ assetId: asset.id, accepted: document.getElementById("hapw-exchange-accepted").checked, requestId: operationId("hapw-exchange") }) });
      updateAssetGuide("exchange", "guideFeedbackExchange");
      close(); toast(t("hapwExchangeDone")); renderAssets();
    } catch (err) { toast(err.message, "error"); button.disabled = false; }
  });
}

function openDexHandoff(clip) {
  const pool = clip.dexPool;
  modalRoot.innerHTML = [
    '<div class="modal-backdrop"><section class="modal confirm-modal dex-handoff-modal" role="dialog" aria-modal="true" aria-labelledby="dex-handoff-title">',
    '<div class="confirm-mark">P</div><div class="confirm-content"><p class="eyebrow">CLIP → USDT</p><h2 id="dex-handoff-title">', t("dexHandoffTitle"), '</h2><p>', t("dexHandoffLead"), '</p>',
    '<div class="handoff-route" aria-label="CLIP to USDT"><span>CLIP</span><i>→</i><span>DEX</span><i>→</i><span>USDT</span></div>',
    '<dl class="confirm-summary"><div><dt>', t("dexAvailableBalance"), '</dt><dd>', money(clip.balance), ' CLIP</dd></div><div><dt>', t("indicativeRate"), '</dt><dd>1 CLIP ≈ ', pool.usdtPerClip.toFixed(3), ' USDT</dd></div><div><dt>', t("poolLiquidity"), '</dt><dd>', money(pool.clipReserve), ' CLIP + ', money(pool.usdtReserve), ' USDT</dd></div><div><dt>', t("snapshotTime"), '</dt><dd>', escapeHtml(pool.updatedAt), '</dd></div></dl>',
    '<ol class="handoff-checklist"><li><b>1</b><span>', t("dexHandoffStep1"), '</span></li><li><b>2</b><span>', t("dexHandoffStep2"), '</span></li><li><b>3</b><span>', t("dexHandoffStep3"), '</span></li></ol>',
    '<p class="handoff-notice">', t("dexHandoffNotice"), '</p><div class="confirm-actions"><button class="button" id="dex-handoff-cancel" type="button">', t("cancel"), '</button><a class="button primary" id="dex-handoff-open" href="', safeExternalUrl(clip.dexUrl), '" target="_blank" rel="noopener noreferrer"><span class="button-icon">↗</span>', t("dexHandoffOpen"), '</a></div></div></section></div>'
  ].join("");
  const close = () => { modalRoot.innerHTML = ""; };
  document.getElementById("dex-handoff-cancel").addEventListener("click", close);
  modalRoot.querySelector(".modal-backdrop").addEventListener("click", event => { if (event.target.classList.contains("modal-backdrop")) close(); });
  const open = document.getElementById("dex-handoff-open");
  open.focus();
  open.addEventListener("click", () => {
    updateAssetGuide("dex", "guideFeedbackDex");
    setTimeout(() => {
      close();
      toast(t("dexHandoffStarted"));
      if (state.route === "/assets") renderAssets();
    }, 0);
  });
}

async function renderStudio() {
  const data = await api("/api/studio");
  const activeRedemptions = data.redemptions.filter(item => item.status === "有效" && item.creditsRemaining > 0);
  const eligibleAssets = activeRedemptions.map(item => data.assets.find(asset => asset.id === item.assetId)).filter(Boolean);
  if (!eligibleAssets.some(item => item.id === state.selectedAsset)) state.selectedAsset = eligibleAssets[0] ? eligibleAssets[0].id : "";
  const selectedRedemption = activeRedemptions.find(item => item.assetId === state.selectedAsset);
  const qualityFactor = state.generationQuality === "pro" ? data.generationAccount.proFactor : data.generationAccount.standardFactor;
  const estimatedCost = Math.ceil(state.generationDuration * qualityFactor);
  const estimatedClipCost = Math.ceil(estimatedCost * data.generationAccount.clipCostPerCredit);
  const stats = [["creditsAvailable", data.generationAccount.balance, t("creditsUnit")], ["clipBalance", data.clip.balance, "CLIP"], ["creditsUsed", data.generationAccount.lifetimeUsed, t("creditsUnit")], ["clipSpent", data.generationAccount.lifetimeClipSpent, "CLIP"]].map(item => '<article class="studio-stat"><span>' + t(item[0]) + '</span><strong>' + money(item[1]) + '</strong><small>' + item[2] + '</small></article>').join("");
  const flow = ["studioFlow1", "studioFlow2", "studioFlow3", "studioFlow4", "studioFlow5"].map((key, index) => '<div><b>0' + (index + 1) + '</b><span>' + t(key) + '</span></div>').join('<i>→</i>');
  const materialRows = data.assets.map(asset => {
    const redemption = data.redemptions.find(item => item.assetId === asset.id);
    let action = '<span class="status pending">' + t("rightsReview") + '</span>';
    if (asset.redemptionStatus === "available") action = '<button class="button secondary" type="button" data-redeem="' + escapeHtml(asset.id) + '">' + t("redeemAction") + '</button>';
    if (asset.redemptionStatus === "redeemed") action = '<div class="redeemed-note"><span class="status">' + t("redeemed") + '</span><small>' + t("licenseReceipt") + ' · ' + (redemption ? escapeHtml(redemption.receipt) : "-") + '</small></div>';
    return '<article class="material-row"><div class="material-id"><span>W</span><div><strong>HAPW ' + escapeHtml(asset.tokenId) + ' · ' + local(asset, "name") + '</strong><small>' + t("rightsHolder") + ' · ' + local(asset, "rightsHolder") + '</small></div></div><p>' + local(asset, "authorizationScope") + '</p><div class="material-yield"><small>' + t("redemptionGrant") + '</small><strong>' + asset.creditYield + ' ' + t("creditsUnit") + ' + ' + Math.floor(asset.creditYield * data.generationAccount.clipGrantPerCredit) + ' CLIP</strong></div>' + action + '</article>';
  }).join("");
  const assetOptions = eligibleAssets.length ? eligibleAssets.map(asset => {
    const redemption = activeRedemptions.find(item => item.assetId === asset.id);
    return '<option value="' + escapeHtml(asset.id) + '" ' + (asset.id === state.selectedAsset ? "selected" : "") + '>HAPW ' + escapeHtml(asset.tokenId) + ' · ' + local(asset, "name") + ' · ' + redemption.creditsRemaining + ' ' + t("creditsUnit") + '</option>';
  }).join("") : '<option value="">' + t("noRedeemedAssets") + '</option>';
  const durations = [15, 30, 60].map(value => '<button class="segment ' + (state.generationDuration === value ? "active" : "") + '" type="button" data-generation-duration="' + value + '">' + value + ' ' + t("seconds") + '</button>').join("");
  const qualities = [["standard", "qualityStandard"], ["pro", "qualityPro"]].map(item => '<button class="segment ' + (state.generationQuality === item[0] ? "active" : "") + '" type="button" data-generation-quality="' + item[0] + '">' + t(item[1]) + '</button>').join("");
  const history = data.generations.map(item => {
    const asset = data.assets.find(assetItem => assetItem.id === item.assetId);
    return '<tr><td><strong>' + local(item, "title") + '</strong><small>' + item.duration + ' ' + t("seconds") + ' · ' + t(item.quality === "pro" ? "qualityPro" : "qualityStandard") + '</small></td><td>HAPW ' + (asset ? escapeHtml(asset.tokenId) : "") + '</td><td>' + item.creditsUsed + '</td><td class="amount out">-' + item.clipCost + ' CLIP</td><td>' + money(item.validViews) + '</td><td>' + statusLabel(item) + '</td><td>' + escapeHtml(item.createdAt) + '</td></tr>';
  }).join("");
  shell([
    '<div class="studio-page">', pageHead(t("studioEyebrow"), t("studioTitle"), t("studioLead")),
    '<section class="operation-note"><strong>', t("operationNoteTitle"), '</strong><p>', t("operationNoteText"), '</p><a class="text-link" href="#/transfer">', t("assetExerciseLink"), ' →</a></section>',
    '<section class="studio-flow"><header><p class="eyebrow">WORKFLOW</p><h2>', t("studioFlowTitle"), '</h2></header><div>', flow, '</div></section>',
    '<section class="studio-stats">', stats, '</section>',
    '<section class="studio-materials"><header><h2>', t("redeemTitle"), '</h2><p>', t("redeemLead"), '</p></header><div class="material-list">', materialRows, '</div></section>',
    '<section class="studio-generator"><div class="generator-copy"><p class="eyebrow">AI VIDEO</p><h2>', t("generateTitle"), '</h2><p>', t("generateLead"), '</p><div class="formula-mini"><span>', t("creditFormula"), '</span><small>', t("dualCostFormulaDesc"), '</small></div></div><form id="generation-form" class="generator-form"><label class="field"><span>', t("eligibleMaterial"), '</span><select name="assetId" ', eligibleAssets.length ? "" : "disabled", '>', assetOptions, '</select></label><div class="field"><span class="field-label">', t("videoDuration"), '</span><div class="segmented">', durations, '</div></div><div class="field"><span class="field-label">', t("quality"), '</span><div class="segmented two">', qualities, '</div></div><div class="generation-cost"><span>', t("costEstimate"), '</span><strong>', estimatedCost, ' ', t("creditsUnit"), ' + ', estimatedClipCost, ' CLIP</strong><small>', selectedRedemption ? selectedRedemption.creditsRemaining + ' ' + t("creditsAvailable") + ' · ' + money(data.clip.balance) + ' CLIP' : t("noRedeemedAssets"), '</small></div><label class="check"><input name="accepted" type="checkbox"><span>', t("acceptGeneration"), '</span></label><button class="button primary wide" type="submit" ', eligibleAssets.length ? "" : "disabled", '>', t("generateAction"), '</button></form></section>',
    '<section class="data-section studio-history"><div class="data-head"><h2>', t("generationHistory"), '</h2><a class="text-link" href="#/clip">', t("mechanismDetails"), ' →</a></div><div class="table-wrap"><table><thead><tr><th>', t("video"), '</th><th>HAPW</th><th>', t("creditsCost"), '</th><th>', t("clipCost"), '</th><th>', t("observedViews"), '</th><th>', t("status"), '</th><th>', t("date"), '</th></tr></thead><tbody>', history, '</tbody></table></div></section></div>'
  ].join(""));
  document.querySelectorAll("[data-redeem]").forEach(button => button.addEventListener("click", () => {
    const asset = data.assets.find(item => item.id === button.dataset.redeem);
    if (asset) openRedemptionConfirm(asset);
  }));
  document.querySelectorAll("[data-generation-duration]").forEach(button => button.addEventListener("click", () => { state.generationDuration = Number(button.dataset.generationDuration); renderStudio(); }));
  document.querySelectorAll("[data-generation-quality]").forEach(button => button.addEventListener("click", () => { state.generationQuality = button.dataset.generationQuality; renderStudio(); }));
  const form = document.getElementById("generation-form");
  form.assetId?.addEventListener("change", event => { state.selectedAsset = event.target.value; renderStudio(); });
  form.addEventListener("submit", async event => {
    event.preventDefault();
    const submit = form.querySelector("[type=submit]");
    submit.disabled = true;
    try {
      await api("/api/generations", { method: "POST", body: JSON.stringify({ assetId: form.assetId.value, duration: state.generationDuration, quality: state.generationQuality, accepted: form.accepted.checked, requestId: operationId("generate") }) });
      toast(t("generationDone"));
      renderStudio();
    } catch (err) { toast(err.message, "error"); submit.disabled = false; }
  });
}

async function renderClip() {
  const data = await api("/api/assets");
  const clip = data.clip;
  const pool = clip.dexPool;
  const exchangePolicy = clip.hapwExchangePolicy || { dailyLimit: 2, usedToday: 0, remainingToday: 2, inventoryTotal: 0, inventoryAvailable: 0 };
  const metrics = [["clipMetricHoldings", data.stats.holdings], ["clipMetricCredits", money(data.generationAccount.balance)], ["clipMetricGrantRate", "0.40 CLIP"], ["clipMetricCostRate", "0.20 CLIP"], ["clipMetricDexRate", pool.clipPerUsdt.toFixed(2) + " CLIP"], ["clipMetricBalance", money(clip.balance) + " CLIP"]].map(item => '<div class="clip-metric"><small>' + t(item[0]) + '</small><strong>' + item[1] + '</strong></div>').join("");
  const plays = [["clipPlay1Title", "clipPlay1Desc"], ["clipPlay2Title", "clipPlay2Desc"], ["clipPlay3Title", "clipPlay3Desc"], ["clipPlay4Title", "clipPlay4Desc"]].map((item, index) => '<article class="clip-play"><span>' + (index + 1) + '</span><div><h3>' + t(item[0]) + '</h3><p>' + t(item[1]) + '</p></div></article>').join("");
  const flow = [["01", "clipFlowHold", "clipFlowHoldDesc"], ["02", "clipFlowRedeem", "clipFlowRedeemDesc"], ["03", "clipFlowGrant", "clipFlowGrantDesc"], ["04", "clipFlowGenerate", "clipFlowGenerateDesc"]].map(item => [
    '<article class="mechanism-node"><b>', item[0], '</b><h3>', t(item[1]), '</h3><p>', t(item[2]), '</p></article>'
  ].join("")).join('<span class="mechanism-arrow">→</span>');
  const exchange = ["exchange1", "exchange2", "exchange3", "exchange4"].map((key, index) => '<li><b>0' + (index + 1) + '</b><span>' + t(key) + '</span></li>').join("");
  const policies = [["supplyPolicyTitle", "supplyPolicyText"], ["usagePolicyTitle", "usagePolicyText"], ["recyclePolicyTitle", "recyclePolicyText"], ["disclosurePolicyTitle", "disclosurePolicyText"]].map((item, index) => [
    '<div class="policy-row"><b>0', index + 1, '</b><h3>', t(item[0]), '</h3><p>', t(item[1]), '</p></div>'
  ].join("")).join("");
  shell([
    '<div class="clip-page"><a class="back-link" href="#/assets">← ', t("backAssets"), '</a>',
    '<section class="clip-hero"><div><p class="eyebrow">', t("clipPageEyebrow"), '</p><h1>', t("clipPageTitle"), '</h1><p class="lede">', t("clipPageLead"), '</p><div class="hero-actions"><a class="button primary" href="#/studio">', t("openStudio"), ' →</a><a class="button" href="', safeExternalUrl(clip.dexUrl), '" target="_blank" rel="noopener noreferrer">', t("dexButton"), ' ↗</a></div></div>',
    '<aside class="clip-balance-sheet"><span class="sheet-mark">P</span><p>', t("clipCurrentBalance"), '</p><strong>', money(clip.balance), ' <small>CLIP</small></strong><dl><div><dt>', t("generationCredits"), '</dt><dd>', money(data.generationAccount.balance), ' ', t("creditsUnit"), '</dd></div><div><dt>', t("indicativeRate"), '</dt><dd>1 USDT ≈ ', pool.clipPerUsdt.toFixed(2), ' CLIP</dd></div><div><dt>', t("poolLiquidity"), '</dt><dd>', money(pool.clipReserve), ' CLIP + ', money(pool.usdtReserve), ' USDT</dd></div><div><dt>', t("snapshotTime"), '</dt><dd>', escapeHtml(pool.updatedAt), '</dd></div></dl></aside></section>',
    '<section class="mechanism-section"><header class="home-section-head"><div><p class="eyebrow">FLOW</p><h2>', t("clipMechanismTitle"), '</h2></div></header><div class="mechanism-flow">', flow, '</div></section>',
    '<section class="clip-numbers"><header><p class="eyebrow">LEDGER SNAPSHOT</p><h2>', t("clipNumbersTitle"), '</h2><p>', t("clipNumbersLead"), '</p></header><div class="clip-metrics">', metrics, '</div></section>',
    '<section class="clip-playbook"><header><p class="eyebrow">USE CASES</p><h2>', t("clipPlayTitle"), '</h2><p>', t("clipPlayLead"), '</p></header><div class="clip-play-grid">', plays, '</div></section>',
    '<section class="relation-section"><div class="relation-title"><p class="eyebrow">HAPW × CREDIT × CLIP</p><h2>', t("clipRelationTitle"), '</h2></div><div class="relation-list"><article><span>W</span><div><h3>', t("hapwMechanismTitle"), '</h3><p>', t("hapwMechanismText"), '</p></div></article><article><span>C</span><div><h3>', t("quotaMechanismTitle"), '</h3><p>', t("quotaMechanismText"), '</p></div></article><article><span>P</span><div><h3>', t("clipMechanismRoleTitle"), '</h3><p>', t("clipMechanismRoleText"), '</p></div></article></div></section>',
    '<section class="formula-section"><header><p class="eyebrow">FORMULA</p><h2>', t("formulaTitle"), '</h2><p>', t("formulaLead"), '</p></header><div class="formula-grid"><article><span>C</span><h3>', t("creditFormulaTitle"), '</h3><strong>', t("creditFormula"), '</strong><p>', t("creditFormulaDesc"), '</p></article><article><span>G</span><h3>', t("rewardFormulaTitle"), '</h3><strong>', t("rewardFormula"), '</strong><p>', t("rewardFormulaDesc"), '</p></article><article><span>P</span><h3>', t("splitFormulaTitle"), '</h3><strong>', t("splitFormula"), '</strong><p>', t("splitFormulaDesc"), '</p></article></div><div class="settlement-rules"><h3>', t("settlementRulesTitle"), '</h3><ol>', ["settlement1", "settlement2", "settlement3", "settlement4"].map((key, index) => '<li><b>0' + (index + 1) + '</b><span>' + t(key) + '</span></li>').join(""), '</ol></div></section>',
    '<section class="exchange-section"><div class="exchange-copy"><p class="eyebrow">USDT ↔ CLIP</p><h2>', t("clipExchangeTitle"), '</h2><p>', t("clipExchangeLead"), '</p><div class="market-note"><strong>', t("marketPrice"), '</strong><span>', t("marketPriceDesc"), '</span></div><div class="pool-snapshot"><div><small>', t("indicativeRate"), '</small><strong>1 CLIP ≈ ', pool.usdtPerClip.toFixed(3), ' USDT</strong></div><div><small>', t("poolLiquidity"), '</small><strong>', money(pool.totalLiquidityUsdt), ' USDT</strong></div></div></div><ol class="exchange-steps">', exchange, '</ol></section>',
    '<section class="reverse-exchange"><div><p class="eyebrow">CLIP → HAPW</p><h2>', t("reverseExchangeTitle"), '</h2><p>', t("reverseExchangeLead"), '</p><div class="reverse-policy"><div><small>', t("dailyLimit"), '</small><strong>', exchangePolicy.dailyLimit, ' HAPW / ', t("day"), '</strong></div><div><small>', t("dailyRemaining"), '</small><strong>', exchangePolicy.remainingToday, '</strong></div><div><small>', t("reserveAvailable"), '</small><strong>', exchangePolicy.inventoryAvailable, ' / ', exchangePolicy.inventoryTotal, '</strong></div></div></div><a class="button secondary" href="#/assets">', t("viewReserve"), ' →</a></section>',
    '<section class="policy-section"><header><p class="eyebrow">POLICY NOTE</p><h2>', t("clipPolicyTitle"), '</h2></header><div class="policy-ledger">', policies, '</div></section>',
    '<section class="clip-warning"><span>!</span><p>', t("clipDemoNotice"), '</p></section></div>'
  ].join(""));
}

function assetOptions(assets) {
  if (!state.selectedAsset) {
    const available = assets.find(item => item.transferable);
    state.selectedAsset = available ? available.id : "";
  }
  return assets.map(asset => [
    '<button class="asset-option ', state.selectedAsset === asset.id ? "selected" : "", '" type="button" data-asset="', escapeHtml(asset.id), '" ', asset.transferable ? "" : "disabled", '>',
    '<span><strong>HAPW ', escapeHtml(asset.tokenId), '</strong><small>', local(asset, "name"), '</small></span><small>', asset.transferable ? t("redeemable") : t("unavailable"), "</small></button>"
  ].join("")).join("");
}

async function renderConvert() {
  const data = await api("/api/assets");
  const options = assetOptions(data.assets);
  const regions = '<option value="sea">' + t("regionSea") + '</option><option value="europe">' + t("regionEurope") + '</option><option value="global">' + t("regionGlobal") + "</option>";
  const durations = [30, 90, 180].map(day => '<button type="button" class="segment ' + (state.conversionDays === day ? "active" : "") + '" data-days="' + day + '">' + day + " " + t("day") + "</button>").join("");
  shell([
    pageHead(t("convertEyebrow"), t("convertTitle"), t("convertLead")),
    '<section class="workspace"><aside class="panel"><div class="panel-inner"><h2>', t("selectAsset"), '</h2><div class="asset-list">', options,
    '</div></div></aside><section class="panel"><div class="panel-inner"><h2>', t("config"), '</h2><form id="conversion-form" class="form-grid">',
    '<label class="field"><span>', t("region"), '</span><select name="region">', regions, '</select></label><div class="field"><span class="field-label">', t("duration"),
    '</span><div class="segmented">', durations, '</div></div><div class="notice">', t("feeNotice"), '</div><label class="check"><input name="accepted" type="checkbox"><span>', t("acceptLicense"),
    '</span></label><button class="button secondary wide" type="submit">', t("submitConversion"), "</button></form></div></section></section>"
  ].join(""));
  document.querySelectorAll("[data-asset]").forEach(button => button.addEventListener("click", () => { state.selectedAsset = button.dataset.asset; renderConvert(); }));
  document.querySelectorAll("[data-days]").forEach(button => button.addEventListener("click", () => { state.conversionDays = Number(button.dataset.days); renderConvert(); }));
  document.getElementById("conversion-form").addEventListener("submit", async event => {
    event.preventDefault();
    const form = event.currentTarget;
    const submit = form.querySelector("[type=submit]");
    submit.disabled = true;
    try {
      await api("/api/conversions", { method: "POST", body: JSON.stringify({ assetId: state.selectedAsset, region: form.region.value, days: state.conversionDays, accepted: form.accepted.checked, requestId: operationId("convert") }) });
      toast(t("conversionDone"));
      location.hash = "#/assets";
    } catch (err) { toast(err.message, "error"); submit.disabled = false; }
  });
}

async function renderTransfer() {
  const data = await api("/api/assets");
  const available = data.assets.filter(item => item.transferable);
  if (!state.selectedAsset || !available.some(item => item.id === state.selectedAsset)) state.selectedAsset = available[0] ? available[0].id : "";
  if (!data.platforms.some(item => item.code === state.selectedPlatform)) state.selectedPlatform = data.platforms[0].code;
  const selected = available.find(item => item.id === state.selectedAsset);
  const platform = data.platforms.find(item => item.code === state.selectedPlatform);
  const select = available.map(asset => '<option value="' + escapeHtml(asset.id) + '" ' + (asset.id === state.selectedAsset ? "selected" : "") + '>HAPW ' + escapeHtml(asset.tokenId) + " · " + local(asset, "name") + "</option>").join("");
  const platformRows = data.platforms.map(item => '<button type="button" class="exercise-platform ' + (item.code === state.selectedPlatform ? "selected" : "") + '" data-exercise-platform="' + escapeHtml(item.code) + '"><span>' + escapeHtml(item.name.slice(0, 1)) + '</span><div><strong>' + escapeHtml(item.name) + '</strong><small>' + t("platform_" + item.code) + '</small></div><b>' + (item.code === state.selectedPlatform ? "✓" : "→") + '</b></button>').join("");
  shell([
    pageHead(t("transferEyebrow"), t("transferTitle"), t("transferLead")),
    '<section class="operation-compare" aria-labelledby="operation-compare-title"><header><p class="eyebrow">', t("operationCompareEyebrow"), '</p><h2 id="operation-compare-title">', t("operationCompareTitle"), '</h2><p>', t("operationCompareLead"), '</p></header><div class="operation-compare-grid"><article><small>EXTERNAL</small><h3>', t("assetExerciseTitle"), '</h3><p>', t("assetExerciseDesc"), '</p></article><article><small>INTERNAL</small><h3>', t("redemptionTitle"), '</h3><p>', t("redemptionDesc"), '</p><a class="text-link" href="#/studio">', t("viewGeneration"), ' →</a></article></div></section>',
    '<section class="workspace exercise-workspace"><aside class="panel"><div class="panel-inner"><h2>', t("selectPlatform"), '</h2><p class="panel-lead">', t("selectPlatformLead"), '</p><div class="exercise-platforms">', platformRows,
    '</div></div></aside><section class="panel"><div class="panel-inner"><h2>', t("transferDetails"), '</h2><form id="transfer-form" class="form-grid"><label class="field"><span>', t("selectAsset"),
    '</span><select name="assetId">', select, '</select></label><div class="summary-card"><small>', t("asset"), '</small><strong>HAPW ', selected ? selected.tokenId : "", " · ", selected ? local(selected, "name") : "",
    '</strong><div class="kv"><div><small>', t("sourceAccount"), '</small><strong>Clipli Wallet</strong></div><div><small>', t("exerciseTarget"), '</small><strong>', platform ? escapeHtml(platform.name) : "-",
    '</strong></div></div></div><label class="check"><input type="checkbox" name="accepted"><span>', t("acceptOwnership"), '</span></label><div class="notice">', t("transferNotice"),
    '</div><button class="button secondary wide" type="submit">', t("submitTransfer"), "</button></form></div></section></section>"
  ].join(""));
  document.querySelectorAll("[data-exercise-platform]").forEach(button => button.addEventListener("click", () => { state.selectedPlatform = button.dataset.exercisePlatform; renderTransfer(); }));
  document.querySelector("select[name=assetId]").addEventListener("change", event => { state.selectedAsset = event.target.value; renderTransfer(); });
  document.getElementById("transfer-form").addEventListener("submit", async event => {
    event.preventDefault();
    const form = event.currentTarget;
    const submit = form.querySelector("[type=submit]");
    submit.disabled = true;
    try {
      await api("/api/exercises", { method: "POST", body: JSON.stringify({ assetId: state.selectedAsset, platformCode: state.selectedPlatform, accepted: form.accepted.checked, requestId: operationId("exercise") }) });
      toast(t("transferDone"));
      location.hash = "#/assets";
    } catch (err) { toast(err.message, "error"); submit.disabled = false; }
  });
}

async function renderBind() {
  const profile = await api("/api/profile");
  const accountRow = profile.overseasAccount
    ? '<div class="account-row"><span class="account-icon">H</span><div><h3>' + t("overseasAccount") + '</h3><p>' + t("authorizedAccount") + " · " + escapeHtml(profile.overseasAccount) + '</p></div><button class="button danger" id="unbind" type="button">' + t("unbind") + "</button></div>"
    : '<form id="bind-form" class="form-grid"><label class="field"><span>' + t("overseasAccount") + '</span><input type="text" name="account" placeholder="' + t("accountPlaceholder") + '"></label><button class="button primary" type="submit">' + t("bindAction") + "</button></form>";
  shell([
    '<div class="account-shell"><section class="account-card">', pageHead(t("bindEyebrow"), t("bindTitle"), t("bindLead")),
    '<div class="account-content">', accountRow,
    '<div class="account-row"><span class="account-icon">☎</span><div><h3>', t("phone"), '</h3><p>', escapeHtml(profile.phone), '</p></div><span class="status">', t("enabled"), '</span></div>',
    '<div class="benefit"><div class="benefit-head"><div><small>HWF LEVEL ', profile.level, '</small><h2>', t("benefits"), '</h2></div><strong>', money(profile.points), ' <small>', t("points"),
    '</small></strong></div><div class="benefit-grid"><div>', t("priority"), '</div><div>', t("discount"), '</div><div>', t("quota"), "</div></div></div></div></section></div>"
  ].join(""));
  const unbind = document.getElementById("unbind");
  if (unbind) unbind.addEventListener("click", async () => { await api("/api/profile/overseas", { method: "DELETE" }); toast(t("success")); renderBind(); });
  const bindForm = document.getElementById("bind-form");
  if (bindForm) bindForm.addEventListener("submit", async event => {
    event.preventDefault();
    try { await api("/api/profile/overseas", { method: "POST", body: JSON.stringify({ account: event.currentTarget.account.value }) }); toast(t("success")); renderBind(); }
    catch (err) { toast(err.message, "error"); }
  });
}

async function renderWallet() {
  const profile = await api("/api/profile");
  const wallets = [
    ["MetaMask", "MM", "browserWallet"],
    ["Coinbase Wallet", "CB", "browserWallet"],
    ["WalletConnect", "WC", "mobileWallet"],
    ["Venly", "VE", "walletService"]
  ];
  if (!state.selectedWallet && profile.walletProvider) state.selectedWallet = profile.walletProvider;
  const rows = wallets.map(wallet => [
    '<button class="wallet-row ', state.selectedWallet === wallet[0] ? "selected" : "", '" type="button" data-wallet="', wallet[0], '"><span class="wallet-logo">', wallet[1],
    '</span><span><strong>', wallet[0], '</strong><small>', t(wallet[2]), '</small></span><span>', profile.walletProvider === wallet[0] ? t("connected") : "→", "</span></button>"
  ].join("")).join("");
  shell([
    '<div class="account-shell"><section class="account-card">', pageHead(t("walletEyebrow"), t("walletTitle"), t("walletLead")),
    '<div class="account-content"><div class="wallet-list">', rows, '</div><button class="button primary wide" id="connect-wallet" style="margin-top:20px">', t("connectSelected"),
    "</button></div></section></div>"
  ].join(""));
  document.querySelectorAll("[data-wallet]").forEach(button => button.addEventListener("click", () => { state.selectedWallet = button.dataset.wallet; renderWallet(); }));
  document.getElementById("connect-wallet").addEventListener("click", async event => {
    if (!state.selectedWallet) return toast(t("chooseWallet"), "error");
    event.currentTarget.disabled = true;
    try { await api("/api/profile/wallet", { method: "POST", body: JSON.stringify({ provider: state.selectedWallet }) }); toast(t("connected")); renderWallet(); }
    catch (err) { toast(err.message, "error"); event.currentTarget.disabled = false; }
  });
}

async function renderSecurity() {
  const profile = await api("/api/profile");
  const settings = [
    ["walletSign", "walletSign", "walletSignDesc"],
    ["expiryReminder", "expiryReminder", "expiryReminderDesc"]
  ].map(item => [
    '<div class="setting-row"><div><h3>', t(item[1]), '</h3><p>', t(item[2]), '</p></div><button class="toggle" type="button" role="switch" data-setting="', item[0],
    '" aria-checked="', profile.settings[item[0]], '" aria-label="', t(item[1]), '"></button></div>'
  ].join("")).join("");
  const overseas = '<div class="setting-row"><div><h3>' + t("overseasAccount") + '</h3><p>' + (profile.overseasAccount ? escapeHtml(profile.overseasAccount) + " · " + t("justNow") : t("noData")) + '</p></div><a class="button" href="#/bind">' + (profile.overseasAccount ? t("unbind") : t("bindAction")) + "</a></div>";
  shell([
    '<div class="account-shell"><section class="account-card">', pageHead(t("securityEyebrow"), t("securityTitle"), t("securityLead")),
    '<div class="account-content">', settings, overseas, "</div></section></div>"
  ].join(""));
  document.querySelectorAll("[data-setting]").forEach(button => button.addEventListener("click", async () => {
    const key = button.dataset.setting;
    try {
      await api("/api/profile/settings", { method: "PATCH", body: JSON.stringify({ [key]: button.getAttribute("aria-checked") !== "true" }) });
      toast(t("success")); renderSecurity();
    } catch (err) { toast(err.message, "error"); }
  }));
}

function renderAbout() {
  const roles = [["creatorRole", "creatorRoleDesc"], ["distributorRole", "distributorRoleDesc"], ["holderRole", "holderRoleDesc"]].map(item => '<article class="role"><h3>' + t(item[0]) + '</h3><p>' + t(item[1]) + "</p></article>").join("");
  const records = [["record1Title", "record1Desc"], ["record2Title", "record2Desc"], ["record3Title", "record3Desc"]].map(item => '<div class="record-row"><h3>' + t(item[0]) + '</h3><p>' + t(item[1]) + '</p></div>').join("");
  const principles = ["principle1", "principle2", "principle3", "principle4"].map(key => '<li><span>' + t(key) + '</span></li>').join("");
  const workflow = [["aboutWorkflow1Title", "aboutWorkflow1Desc"], ["aboutWorkflow2Title", "aboutWorkflow2Desc"], ["aboutWorkflow3Title", "aboutWorkflow3Desc"]].map(item => '<article><h3>' + t(item[0]) + '</h3><p>' + t(item[1]) + '</p></article>').join("");
  const faq = Array.from({ length: 8 }, (_, index) => '<details class="faq-item" ' + (index === 0 ? "open" : "") + '><summary><span>' + t("faq" + (index + 1) + "Q") + '</span><b>+</b></summary><p>' + t("faq" + (index + 1) + "A") + '</p></details>').join("");
  shell([
    '<div class="about-page">', pageHead(t("aboutEyebrow"), t("aboutTitle"), t("aboutLead")),
    '<section class="about-why"><div><h2>', t("aboutWhyTitle"), '</h2><div class="about-columns"><p>', t("aboutWhyText1"), '</p><p>', t("aboutWhyText2"), '</p></div></div></section>',
    '<section class="about-grid"><article class="about-block accent"><h2>', t("utilityTitle"), '</h2><p>', t("utilityText"),
    '</p><a class="text-link" href="#/assets">', t("manageAssets"), ' →</a></article><article class="about-block mint"><h2>', t("rightsTitle"),
    '</h2><p>', t("rightsText"), '</p><a class="text-link" href="#/clip">', t("mechanismDetails"), ' →</a></article></section>',
    '<section class="record-section"><header><div><h2>', t("recordTitle"), '</h2><p>', t("recordIntro"), '</p></div></header><div class="record-list">', records, '</div></section>',
    '<section class="role-band"><div class="role-heading"><h2 class="section-title">', t("platformRoles"), '</h2></div><div class="role-grid">', roles, '</div></section>',
    '<section class="about-workflow"><header><p class="eyebrow">WORKFLOW</p><h2>', t("aboutWorkflowTitle"), '</h2><p>', t("aboutWorkflowLead"), '</p></header><div class="workflow-grid">', workflow, '</div></section>',
    '<section class="about-boundary"><div><p class="eyebrow">BOUNDARY</p><h2>', t("boundaryTitle"), '</h2><p>', t("boundaryText"), '</p></div><div><p class="eyebrow">PRINCIPLES</p><h2>', t("principleTitle"), '</h2><ol>', principles, '</ol></div></section>',
    '<section class="faq-section"><header><p class="eyebrow">Q&A</p><h2>', t("faqTitle"), '</h2><p>', t("faqLead"), '</p></header><div class="faq-list">', faq, '</div></section>',
    '<section class="about-manifesto"><div><p class="eyebrow">Clipli OPERATING NOTE</p><h2>', t("governanceTitle"), '</h2><p>', t("governanceText"), '</p></div></section></div>'
  ].join(""));
}

async function render() {
  state.route = (location.hash.slice(1) || "/").split("?")[0];
  document.documentElement.lang = { zh: "zh-CN", en: "en", es: "es", ja: "ja", fr: "fr", ko: "ko-KR" }[state.lang] || "zh-CN";
  const routeTitles = {
    "/": "navHome",
    "/works": "navWorks",
    "/work": "details",
    "/assets": "navAssets",
    "/studio": "navStudio",
    "/clip": "clipPageTitle",
    "/convert": "convertTitle",
    "/transfer": "navTransfer",
    "/bind": "bindTitle",
    "/wallet": "walletTitle",
    "/security": "securityTitle",
    "/about": "navAbout"
  };
  const titleKey = state.route.startsWith("/work/") ? "details" : routeTitles[state.route];
  document.title = "Clipli · " + t(titleKey || "navHome");
  const themeColor = document.querySelector('meta[name="theme-color"]');
  if (themeColor) themeColor.content = state.theme === "dark" ? "#11110f" : "#f2efe7";
  loading();
  try {
    if (state.route === "/") return await renderHome();
    if (state.route === "/works") return await renderWorks();
    if (state.route.startsWith("/work/")) return await renderWorkDetail();
    if (state.route === "/assets") return await renderAssets();
    if (state.route === "/studio") return await renderStudio();
    if (state.route === "/clip") return await renderClip();
    if (state.route === "/convert") return await renderConvert();
    if (state.route === "/transfer") return await renderTransfer();
    if (state.route === "/bind") return await renderBind();
    if (state.route === "/wallet") return await renderWallet();
    if (state.route === "/security") return await renderSecurity();
    if (state.route === "/about") return renderAbout();
    location.hash = "#/";
  } catch (err) {
    shell('<div class="empty"><h1>' + t("failed") + '</h1><p>' + escapeHtml(err.message) + "</p></div>");
  }
}

window.addEventListener("hashchange", render);
window.addEventListener("keydown", event => {
  if (event.key === "Escape" && modalRoot.innerHTML) modalRoot.innerHTML = "";
});
render();
