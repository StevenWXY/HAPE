// Shared platform and real-asset copy for all supported interface languages.
const platformCopy = {
  navPlatforms: ['绑定', 'Link platforms', 'Vincular plataformas', '外部連携', 'Lier une plateforme', '외부 플랫폼 연결'],
  navPlatformsShort: ['绑定', 'Apps', 'Apps', '連携', 'Apps', '연결'],
  navWalletShort: ['钱包', 'Wallet', 'Cartera', '財布', 'Wallet', '지갑'],
  bindTitle: ['外部资产平台', 'External asset platforms', 'Plataformas de activos', '外部資産プラットフォーム', 'Plateformes d’actifs', '외부 자산 플랫폼'],
  bindLead: ['绑定你已有的平台账号，核对持仓与划转状态。海文发支持短信验证绑定。', 'Link your existing account and review holdings and transfers. HAIWEN supports SMS verification.', 'Vincula tu cuenta y revisa posiciones y transferencias. HAIWEN admite verificación SMS.', '既存のアカウントを連携し、保有資産と移管状況を確認します。海文発はSMS認証に対応しています。', 'Liez votre compte et consultez vos actifs et transferts. HAIWEN utilise une vérification SMS.', '기존 계정을 연결하고 보유 자산과 이전 상태를 확인하세요. 해문발은 SMS 인증을 지원합니다.'],
  realAssetsOnly: ['资产均来自外部平台的真实划转', 'Assets come from confirmed external transfers', 'Activos procedentes de transferencias externas confirmadas', '資産は確認済みの外部移管から取得', 'Actifs issus de transferts externes confirmés', '확인된 외부 이전으로 확보한 자산'],
  connectFirstPlatform: ['绑定', 'Link a platform to get started', 'Vincular una plataforma', 'プラットフォームを連携', 'Lier une plateforme', '플랫폼 연결하기'],
  noAssetsYet: ['还没有划转到 Clipli 的资产', 'No assets transferred to Clipli yet', 'Aún no hay activos transferidos a Clipli', 'Clipliへの移管済み資産はありません', 'Aucun actif transféré vers Clipli', 'Clipli로 이전된 자산이 없습니다'],
  noEligibleAssets: ['暂无可用于此操作的资产', 'No eligible assets for this action', 'No hay activos aptos para esta acción', 'この操作に利用できる資産はありません', 'Aucun actif éligible pour cette action', '이 작업에 사용할 자산이 없습니다'],
  assetEmptyLead: ['先绑定海文发等外部平台并完成资产划转。绑定账号或连接钱包不会自动增加资产。', 'Link an external platform and complete a transfer first. Linking an account or wallet does not add assets.', 'Vincula una plataforma y completa una transferencia. Conectar cuentas o carteras no añade activos.', '外部プラットフォームの連携と資産移管を完了してください。アカウントやウォレットの接続だけでは資産は増えません。', 'Liez une plateforme puis terminez un transfert. Lier un compte ou un portefeuille n’ajoute aucun actif.', '외부 플랫폼을 연결하고 자산 이전을 완료하세요. 계정이나 지갑 연결만으로 자산이 추가되지 않습니다.'],
  noRecordsYet: ['暂无记录', 'No records yet', 'Sin registros', '記録はありません', 'Aucun enregistrement', '기록이 없습니다'],
  noWorksYet: ['暂无已接入的作品', 'No works available yet', 'Aún no hay obras disponibles', '公開作品はありません', 'Aucune œuvre disponible', '등록된 작품이 없습니다'],
  confirmedPortfolio: ['已划转资产', 'Transferred assets', 'Activos transferidos', '移管済み資産', 'Actifs transférés', '이전된 자산'],
  confirmedAssetsLead: ['仅计入已确认划转的资产，每笔保留来源平台与划转凭据。', 'Only confirmed transfers are included, with the source platform and transfer receipt.', 'Solo incluye transferencias confirmadas con plataforma de origen y comprobante.', '確認済みの移管のみを計上し、移管元と証憑を保存します。', 'Seuls les transferts confirmés sont comptabilisés, avec leur origine et justificatif.', '확인된 이전만 집계하며 출처 플랫폼과 이전 증빙을 보관합니다.'],
  balanceFromRecords: ['余额以实际入账与使用记录为准。', 'Balance reflects actual ledger entries and usage.', 'El saldo refleja los registros y usos reales.', '残高は実際の入出金と利用記録に基づきます。', 'Le solde reflète les écritures et utilisations réelles.', '잔액은 실제 입출금 및 사용 기록을 반영합니다.'],
  assetSource: ['来源平台', 'Source platform', 'Plataforma de origen', '移管元', 'Plateforme d’origine', '출처 플랫폼'],
  transferReceipt: ['划转凭据', 'Transfer receipt', 'Comprobante', '移管証憑', 'Justificatif de transfert', '이전 증빙'],
  reserveRealLead: ['仅展示已从外部平台入库的可兑换资产。', 'Only externally transferred reserve assets are available.', 'Solo se ofrecen reservas recibidas de plataformas externas.', '外部から移管済みの交換用資産のみを表示します。', 'Seules les réserves transférées depuis l’extérieur sont proposées.', '외부에서 이전된 교환 가능 자산만 표시합니다.'],
  noReserveYet: ['暂无可兑换的真实储备资产', 'No reserve assets available', 'No hay activos de reserva disponibles', '交換可能な準備資産はありません', 'Aucun actif de réserve disponible', '교환 가능한 준비 자산이 없습니다'],
  includesFee: ['含 5% 手续费', 'Includes 5% fee', 'Incluye comisión del 5 %', '手数料5%込み', 'Frais de 5 % inclus', '수수료 5% 포함'],
  integrationComing: ['暂未接入', 'Not connected yet', 'Aún no integrado', '未対応', 'Pas encore intégré', '아직 미지원'],
  openseaComingLead: ['OpenSea 绑定与资产划转暂未开放；连接钱包不会自动同步 OpenSea 资产。', 'OpenSea linking and transfers are not available yet. Connecting a wallet does not sync OpenSea assets.', 'La vinculación y transferencias de OpenSea aún no están disponibles. Conectar una cartera no sincroniza activos.', 'OpenSea連携と移管は準備中です。ウォレット接続ではOpenSea資産は同期されません。', 'La liaison et les transferts OpenSea sont indisponibles. Connecter un portefeuille ne synchronise pas ces actifs.', 'OpenSea 연결 및 이전은 준비 중입니다. 지갑을 연결해도 OpenSea 자산은 동기화되지 않습니다.'],
  platformComingLead: ['账号绑定与资产划转尚未开放。', 'Account linking and transfers are not available yet.', 'La vinculación y transferencias aún no están disponibles.', 'アカウント連携と移管は準備中です。', 'La liaison et les transferts ne sont pas encore disponibles.', '계정 연결 및 이전은 준비 중입니다.'],
  visitPlatform: ['访问官网', 'Visit website', 'Visitar sitio', '公式サイト', 'Site officiel', '공식 사이트'],
  haiwenUserId: ['海文发用户 ID', 'HAIWEN user ID', 'ID de usuario HAIWEN', '海文発ユーザーID', 'Identifiant HAIWEN', '해문발 사용자 ID'],
  haiwenUserIdHelp: ['填写海文发账号对应的用户 ID，可在海文发账户信息中确认。', 'Use the user ID associated with your HAIWEN account.', 'Usa el ID asociado a tu cuenta HAIWEN.', '海文発アカウントに対応するユーザーIDを入力してください。', 'Utilisez l’identifiant associé à votre compte HAIWEN.', '해문발 계정에 해당하는 사용자 ID를 입력하세요.'],
  bindingPhone: ['海文发账号手机号', 'HAIWEN phone number', 'Teléfono de HAIWEN', '海文発登録電話番号', 'Téléphone HAIWEN', '해문발 계정 전화번호'],
  smsCode: ['短信验证码', 'SMS code', 'Código SMS', 'SMS認証コード', 'Code SMS', 'SMS 인증번호'],
  sendCode: ['发送验证码', 'Send code', 'Enviar código', 'コードを送信', 'Envoyer le code', '인증번호 발송'],
  resendCode: ['重新发送', 'Resend', 'Reenviar', '再送信', 'Renvoyer', '재발송'],
  acceptPlatformBinding: ['我同意绑定此海文发账号，并授权 Clipli 查询其资产持仓；绑定本身不会核销或划转资产。', 'I agree to link this HAIWEN account and allow Clipli to read its holdings. Linking does not redeem or transfer assets.', 'Acepto vincular esta cuenta y consultar sus activos. Vincular no canjea ni transfiere activos.', '海文発との連携とClipliによる保有資産の照会に同意します。連携だけで償却や移管は行われません。', 'J’accepte de lier ce compte et de permettre la consultation de ses actifs, sans rachat ni transfert automatique.', '해문발 계정을 연결하고 Clipli의 자산 조회를 허용합니다. 연결만으로 자산이 소각되거나 이전되지 않습니다.'],
  verifyAndBind: ['绑定', 'Verify and link HAIWEN', 'Verificar y vincular HAIWEN', '認証して海文発を連携', 'Vérifier et lier HAIWEN', '인증 후 해문발 연결'],
  platformBound: ['已绑定', 'Linked', 'Vinculado', '連携済み', 'Lié', '연결됨'],
  boundAt: ['绑定时间', 'Linked at', 'Fecha de vinculación', '連携日時', 'Date de liaison', '연결 시각'],
  bindingDoesNotTransfer: ['外部持仓与 Clipli 资产分别显示。只有外部平台确认划转成功，才会计入 Clipli 资产。', 'External holdings and Clipli assets are separate. Assets enter Clipli only after the external platform confirms the transfer.', 'Las posiciones externas se muestran aparte. Solo se incorporan a Clipli tras confirmar la transferencia.', '外部保有とClipli資産は別に表示します。外部で移管が確認された後にClipliへ計上されます。', 'Les actifs externes sont affichés séparément et entrent dans Clipli après confirmation du transfert.', '외부 보유 자산과 Clipli 자산은 별도로 표시됩니다. 외부 플랫폼이 이전을 확인한 뒤 Clipli에 반영됩니다.'],
  transferHelp: ['当前海文发划转由平台审核处理。请联系平台确认资产与数量，完成后在“已划转资产”中刷新查看。', 'HAIWEN transfers currently require platform review. Contact the platform to confirm assets and quantities, then refresh Transferred assets after completion.', 'Las transferencias HAIWEN requieren revisión. Confirma activos y cantidades con la plataforma y actualiza tras completarlas.', '海文発の移管は現在プラットフォーム審査が必要です。対象と数量を確認し、完了後に移管済み資産を更新してください。', 'Les transferts HAIWEN nécessitent une validation. Confirmez les actifs et quantités avec la plateforme, puis actualisez après confirmation.', '해문발 이전은 플랫폼 검토가 필요합니다. 플랫폼에 자산과 수량을 확인하고 완료 후 이전된 자산을 새로고침하세요.'],
  viewExternalHoldings: ['查看海文发持仓', 'View HAIWEN holdings', 'Ver activos HAIWEN', '海文発の保有を照会', 'Voir les actifs HAIWEN', '해문발 보유 자산 조회'],
  bindingStep1: ['验证并绑定账号', 'Verify and link', 'Verificar y vincular', '認証して連携', 'Vérifier et lier', '인증 및 연결'],
  bindingStep2: ['核对外部持仓', 'Review external holdings', 'Revisar activos externos', '外部保有を確認', 'Vérifier les actifs externes', '외부 보유 자산 확인'],
  bindingStep3: ['平台确认划转后入账', 'Credit after confirmed transfer', 'Registrar tras confirmación', '移管確認後に計上', 'Créditer après confirmation', '이전 확인 후 반영'],
  haiwenReadyLead: ['使用海文发短信验证，关联现有账号。', 'Link your existing account with HAIWEN SMS verification.', 'Vincula tu cuenta con verificación SMS de HAIWEN.', '海文発のSMS認証で既存アカウントを連携します。', 'Liez votre compte existant par SMS HAIWEN.', '해문발 SMS 인증으로 기존 계정을 연결합니다.'],
  refreshStatus: ['刷新状态', 'Refresh status', 'Actualizar estado', '状態を更新', 'Actualiser', '상태 새로고침'],
  bindingStatusUnknown: ['暂时无法确认绑定状态，请刷新重试。', 'Unable to confirm the link status. Refresh to retry.', 'No se puede confirmar la vinculación. Actualiza para reintentar.', '連携状態を確認できません。更新して再試行してください。', 'Impossible de confirmer la liaison. Actualisez pour réessayer.', '연결 상태를 확인할 수 없습니다. 새로고침 후 다시 시도하세요.'],
  connectionUnavailable: ['海文发连接暂不可用', 'HAIWEN connection unavailable', 'Conexión HAIWEN no disponible', '海文発に接続できません', 'Connexion HAIWEN indisponible', '해문발 연결을 사용할 수 없습니다'],
  connectionUnavailableLead: ['当前服务尚未启用海文发连接，请联系平台开通后刷新。你的账号不会被自动绑定。', 'The service has not enabled HAIWEN yet. Contact the platform and refresh once enabled.', 'El servicio aún no ha habilitado HAIWEN. Contacta con la plataforma y actualiza después.', 'このサービスでは海文発接続が有効化されていません。運営に確認後、更新してください。', 'La connexion HAIWEN n’est pas activée. Contactez la plateforme puis actualisez.', '서비스에서 해문발 연결이 활성화되지 않았습니다. 플랫폼에 문의 후 새로고침하세요.'],
  otherPlatforms: ['其他资产平台', 'Other asset platforms', 'Otras plataformas', 'その他の資産プラットフォーム', 'Autres plateformes', '기타 자산 플랫폼'],
  sendingCode: ['正在发送验证码…', 'Sending code…', 'Enviando código…', '送信中…', 'Envoi du code…', '인증번호 발송 중…'],
  codeSent: ['验证码已发送至', 'Code sent to', 'Código enviado a', 'コード送信先：', 'Code envoyé au', '인증번호 발송 완료:'],
  verifyingBinding: ['正在验证并绑定…', 'Verifying and linking…', 'Verificando y vinculando…', '認証・連携中…', 'Vérification et liaison…', '인증 및 연결 중…'],
  loadingExternalHoldings: ['正在查询海文发持仓…', 'Loading HAIWEN holdings…', 'Consultando activos HAIWEN…', '海文発保有を照会中…', 'Chargement des actifs HAIWEN…', '해문발 보유 자산 조회 중…'],
  externalHoldingsTitle: ['海文发持仓查询', 'HAIWEN holdings', 'Activos HAIWEN', '海文発の保有資産', 'Actifs HAIWEN', '해문발 보유 자산'],
  externalHoldingsLead: ['按当前页资产查询持有数量；“—”表示平台未返回数量。此列表不计入 Clipli 资产。', 'Holdings are queried for this page. “—” means no count was returned. This list is not part of your Clipli portfolio.', 'Se consultan los activos de esta página. «—» indica cantidad no recibida. No forman parte de tu cartera Clipli.', 'このページの資産の保有数を照会します。「—」は未回答です。この一覧はClipli残高には含まれません。', 'Les quantités concernent cette page. « — » signifie non communiqué. Cette liste est distincte du portefeuille Clipli.', '현재 페이지의 보유 수량을 조회합니다. “—”는 수량 미응답입니다. 이 목록은 Clipli 자산에 포함되지 않습니다.'],
  externalQuantity: ['外部持有数量', 'External quantity', 'Cantidad externa', '外部保有数', 'Quantité externe', '외부 보유 수량'],
  transferAvailability: ['划转状态', 'Transfer availability', 'Disponibilidad de transferencia', '移管対応状況', 'Disponibilité du transfert', '이전 가능 상태'],
  transferByPlatform: ['需平台审核划转', 'Platform review required', 'Requiere revisión', '運営審査が必要', 'Validation requise', '플랫폼 검토 필요'],
  mappingPending: ['暂不支持划转', 'Transfer not available yet', 'Transferencia no disponible', '移管は未対応', 'Transfert indisponible', '이전 미지원'],
  previousPage: ['上一页', 'Previous page', 'Página anterior', '前のページ', 'Page précédente', '이전 페이지'],
  nextPage: ['下一页', 'Next page', 'Página siguiente', '次のページ', 'Page suivante', '다음 페이지'],
  marketUnavailable: ['行情与兑换尚未开放', 'Market data and swap unavailable', 'Mercado y canje no disponibles', '相場・交換は準備中', 'Cours et échange indisponibles', '시세 및 교환 미지원'],
  error_verification_expired: ['验证码已过期，请重新发送。', 'Code expired. Request a new one.', 'Código caducado. Solicita otro.', 'コードが期限切れです。再送信してください。', 'Code expiré. Demandez-en un autre.', '인증번호가 만료되었습니다. 다시 발송하세요.'],
  error_verification_code_invalid: ['验证码不正确，请检查后重试。', 'Incorrect code. Check and retry.', 'Código incorrecto. Comprueba y reintenta.', 'コードが正しくありません。確認してください。', 'Code incorrect. Vérifiez et réessayez.', '인증번호가 틀립니다. 확인 후 다시 시도하세요.'],
  error_external_platform_not_configured: ['海文发连接尚未启用，请联系平台。', 'HAIWEN is not enabled. Contact the platform.', 'HAIWEN no está habilitado. Contacta con la plataforma.', '海文発接続が無効です。運営に確認してください。', 'HAIWEN n’est pas activé. Contactez la plateforme.', '해문발 연결이 활성화되지 않았습니다. 플랫폼에 문의하세요.'],
  error_external_platform_unavailable: ['海文发暂时无法连接，请稍后重试。', 'HAIWEN is unavailable. Try again later.', 'HAIWEN no está disponible. Reintenta más tarde.', '海文発に接続できません。後で再試行してください。', 'HAIWEN est indisponible. Réessayez plus tard.', '해문발에 연결할 수 없습니다. 나중에 다시 시도하세요.'],
  error_external_rate_limited: ['操作过于频繁，请稍后重试。', 'Too many requests. Try again later.', 'Demasiadas solicitudes. Reintenta más tarde.', '操作が多すぎます。後で再試行してください。', 'Trop de demandes. Réessayez plus tard.', '요청이 너무 많습니다. 나중에 다시 시도하세요.'],
  error_binding_conflict: ['此账号已有关联，请刷新绑定状态后核对。', 'This account is already linked. Refresh to check.', 'La cuenta ya está vinculada. Actualiza para comprobar.', '既に連携されています。状態を更新してください。', 'Ce compte est déjà lié. Actualisez pour vérifier.', '이미 연결된 계정입니다. 상태를 새로고침하세요.']
};
['zh', 'en', 'es', 'ja', 'fr', 'ko'].forEach((lang, index) => {
  Object.entries(platformCopy).forEach(([key, values]) => { window.CLIPLI_LOCALES[lang][key] = values[index]; });
});
const partnerErrors = {
  error_external_invalid_request: ['海文发未接受此请求，请检查用户 ID、手机号和验证码。', 'HAIWEN rejected the request. Check the user ID, phone and code.', 'HAIWEN rechazó la solicitud. Revisa ID, teléfono y código.', '海文発が要求を受理しませんでした。ID・電話番号・コードを確認してください。', 'HAIWEN a refusé la demande. Vérifiez l’identifiant, le téléphone et le code.', '해문발이 요청을 거절했습니다. ID, 전화번호 및 인증번호를 확인하세요.'],
  error_external_credentials_invalid: ['平台连接认证异常，请联系平台处理。', 'Platform connection authentication failed. Contact support.', 'Falló la autenticación de la conexión. Contacta con soporte.', 'プラットフォーム接続の認証エラーです。運営へお問い合わせください。', 'Échec d’authentification de la connexion. Contactez l’assistance.', '플랫폼 연결 인증 오류입니다. 운영자에게 문의하세요.'],
  error_external_not_found: ['海文发未找到此账号或资产，请核对后重试。', 'HAIWEN could not find this account or asset. Check and retry.', 'HAIWEN no encontró la cuenta o activo. Comprueba y reintenta.', 'アカウントまたは資産が見つかりません。確認して再試行してください。', 'Compte ou actif introuvable sur HAIWEN. Vérifiez et réessayez.', '해문발에서 계정 또는 자산을 찾을 수 없습니다. 확인 후 다시 시도하세요.'],
  error_external_conflict: ['账号或手机号存在绑定冲突，请联系平台核对。', 'The account or phone has a binding conflict. Contact the platform.', 'Existe un conflicto de vinculación. Contacta con la plataforma.', 'アカウントまたは電話番号の連携が競合しています。運営に確認してください。', 'Conflit de liaison du compte ou du téléphone. Contactez la plateforme.', '계정 또는 전화번호 연결 충돌입니다. 플랫폼에 문의하세요.'],
  error_verification_required: ['请先发送验证码，再完成绑定。', 'Request a verification code before linking.', 'Solicita un código antes de vincular.', '先に認証コードを送信してください。', 'Demandez un code avant de lier le compte.', '먼저 인증번호를 발송한 후 연결하세요.']
};
['zh', 'en', 'es', 'ja', 'fr', 'ko'].forEach((lang, index) => {
  Object.entries(partnerErrors).forEach(([key, values]) => { window.CLIPLI_LOCALES[lang][key] = values[index]; });
  const copy = window.CLIPLI_LOCALES[lang];
  ['invalid_external_user_id', 'invalid_phone'].forEach(code => { copy['error_' + code] = copy.error_external_invalid_request; });
  copy.error_invalid_verification_code = copy.error_verification_code_invalid;
  ['external_binding_mismatch', 'external_binding_invalid', 'external_binding_not_confirmed'].forEach(code => { copy['error_' + code] = copy.bindingStatusUnknown; });
});

const sandboxCopy = {
 sandboxConnection: ['测试环境', 'Test environment', 'Entorno de pruebas', 'テスト環境', 'Environnement de test', '테스트 환경'],
 sandboxTransferHelp: ['当前连接海文发测试环境，可联调绑定和持仓查询。正式资产划转暂未开放，测试数据不会计入 Clipli 资产。', 'HAIWEN is connected to a test environment for binding and holdings checks. Real transfers are disabled; test data never enters the Clipli portfolio.', 'HAIWEN usa un entorno de pruebas. Las transferencias reales están desactivadas y los datos de prueba no se incorporan a Clipli.', '海文発のテスト環境で連携・保有照会を確認できます。正式移管は無効で、テストデータはClipli資産に計上されません。', 'HAIWEN est connecté en test. Les transferts réels sont désactivés et les données de test ne sont pas ajoutées au portefeuille.', '해문발 테스트 환경에서 연결 및 보유 조회를 확인할 수 있습니다. 실제 이전은 비활성화되며 테스트 데이터는 Clipli 자산에 반영되지 않습니다.'],
 sandboxHoldingsLead: ['以下为海文发测试环境返回的数据，仅供联调，不代表真实持仓，也不会计入 Clipli 资产。', 'These holdings come from the HAIWEN test environment. They are test data, not real holdings, and are excluded from Clipli assets.', 'Estos datos proceden del entorno de pruebas. No representan activos reales ni se incorporan a Clipli.', '海文発テスト環境のデータです。実際の保有資産ではなく、Clipli資産には含まれません。', 'Ces données viennent du test HAIWEN. Elles ne représentent pas des actifs réels et sont exclues de Clipli.', '해문발 테스트 환경 데이터입니다. 실제 보유 자산이 아니며 Clipli 자산에 포함되지 않습니다.']
};
['zh', 'en', 'es', 'ja', 'fr', 'ko'].forEach((lang, index) => {
 Object.entries(sandboxCopy).forEach(([key, values]) => { window.CLIPLI_LOCALES[lang][key] = values[index]; });
 window.CLIPLI_LOCALES[lang].error_external_sandbox_transfer_disabled = sandboxCopy.sandboxTransferHelp[index];
});
