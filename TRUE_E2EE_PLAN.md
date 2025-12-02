# 真正的端到端加密方案

## 问题分析

当前问题：
1. 使用 base64 编码"加密" AES 密钥 → 不是真正的加密
2. 服务器可以解码并获取 AES 密钥 → 违背端到端加密原则
3. 如果攻击者获取加密密钥，可以绕过签名验证 → 应用无意义

## 解决方案：使用钱包公钥的真正非对称加密

### 可用的钱包功能
1. **MetaMask 等现代钱包提供**：
   - `eth_getEncryptionPublicKey(address)` - 获取加密公钥
   - `eth_decrypt(encryptedData, address)` - 使用私钥解密
   - `personal_sign` - 签名消息（已有）

2. **加密标准**：
   - 使用 `x25519-xsalsa20-poly1305` 算法
   - 这是 MetaMask 和许多以太坊钱包使用的标准

### 新的加密流程

#### 1. 获取钱包加密公钥
```javascript
// 请求用户授权获取加密公钥
const encryptionPublicKey = await window.ethereum.request({
  method: 'eth_getEncryptionPublicKey',
  params: [address],
});
```

#### 2. 使用公钥加密 AES 密钥
```javascript
// 使用 eth-sig-util 或类似库
import { encrypt } from 'eth-sig-util';

const encryptedAesKey = encrypt(
  encryptionPublicKey,
  { data: aesKeyString },
  'x25519-xsalsa20-poly1305'
);
```

#### 3. 存储加密数据
- 服务器存储：`{ encryptedData, iv, encryptedAesKey }`
- 服务器**不能**解密 `encryptedAesKey`

#### 4. 解密时请求钱包解密
```javascript
// 用户需要授权解密
const decryptedAesKey = await window.ethereum.request({
  method: 'eth_decrypt',
  params: [encryptedAesKey, address],
});
```

### 优势
1. **真正的端到端加密**：只有用户钱包能解密
2. **服务器零知识**：服务器无法解密任何内容
3. **使用标准算法**：`x25519-xsalsa20-poly1305`
4. **用户控制**：每次解密都需要钱包授权

### 实现步骤

#### 阶段 1：安装依赖
```bash
cd frontend
npm install eth-sig-util @metamask/eth-sig-util
```

#### 阶段 2：修改加密工具
1. 添加钱包加密公钥获取功能
2. 使用 `eth-sig-util` 进行加密
3. 使用钱包的 `eth_decrypt` 进行解密

#### 阶段 3：更新前端页面
1. 修改 `CreateContentPage` 使用新加密
2. 修改 `ContentDetailPage` 使用新解密
3. 添加用户授权提示

#### 阶段 4：测试
1. 测试完整加密/解密流程
2. 验证服务器不能解密
3. 测试错误处理

### 代码示例

#### 新的加密函数
```typescript
import { encrypt } from 'eth-sig-util';

export async function getWalletEncryptionPublicKey(address: string): Promise<string> {
  return await window.ethereum.request({
    method: 'eth_getEncryptionPublicKey',
    params: [address],
  });
}

export async function encryptWithWalletPublicKey(
  data: string,
  publicKey: string
): Promise<string> {
  return encrypt(publicKey, { data }, 'x25519-xsalsa20-poly1305');
}

export async function decryptWithWallet(
  encryptedData: string,
  address: string
): Promise<string> {
  return await window.ethereum.request({
    method: 'eth_decrypt',
    params: [encryptedData, address],
  });
}
```

#### 更新创建内容流程
```typescript
// 1. 获取钱包加密公钥
const encryptionPublicKey = await getWalletEncryptionPublicKey(address);

// 2. 生成 AES 密钥并加密内容
const aesKey = await generateAESKey();
const aesKeyString = await exportAESKey(aesKey);
const { encryptedData, iv } = await encryptWithAES(content, aesKey);

// 3. 使用钱包公钥加密 AES 密钥
const encryptedAesKey = await encryptWithWalletPublicKey(aesKeyString, encryptionPublicKey);

// 4. 发送到服务器
await contentAPI.create(title, encryptedAesKey, iv, encryptedData);
```

#### 更新解密流程
```typescript
// 1. 获取加密数据
const { encryptedAesKey, encryptedData, iv } = await contentAPI.decrypt(...);

// 2. 请求钱包解密 AES 密钥
const aesKeyString = await decryptWithWallet(encryptedAesKey, address);

// 3. 解密内容
const aesKey = await importAESKey(aesKeyString);
const decrypted = await decryptWithAES(encryptedData, aesKey, iv);
```

### 安全考虑

#### 1. 用户授权
- 获取加密公钥需要用户授权
- 每次解密需要用户授权
- 钱包会显示明确的授权对话框

#### 2. 密钥管理
- 私钥永远在钱包中，不在应用代码中
- 没有 localStorage 存储私钥
- 使用硬件钱包更安全

#### 3. 前向保密
- 每次会话使用不同的 AES 密钥
- 即使一个密钥泄露，其他内容仍然安全

#### 4. 审计日志
- 钱包会记录加密/解密请求
- 用户可以查看授权历史

### 兼容性

#### 支持的钱包
1. **MetaMask** - 完全支持
2. **Coinbase Wallet** - 可能支持
3. **WalletConnect** - 需要检查
4. **其他以太坊钱包** - 如果实现相同标准

#### 回退方案
如果钱包不支持加密 API：
1. 使用当前的改进方案（AES-GCM-SHA256）
2. 显示警告给用户
3. 建议使用支持加密的钱包

### 迁移计划

#### 数据迁移
1. 新内容使用新加密方案
2. 旧内容保持兼容
3. 可以逐步重新加密旧内容

#### 版本标识
在加密数据中添加版本标识：
```json
{
  "version": "2.0",
  "algorithm": "x25519-xsalsa20-poly1305",
  "encryptedKey": "...",
  "iv": "...",
  "encryptedData": "..."
}
```

### 结论

这个方案提供了真正的端到端加密：
- ✅ 只有用户能解密自己的数据
- ✅ 服务器零知识
- ✅ 使用标准加密算法
- ✅ 用户控制每次解密
- ✅ 兼容主流钱包

这解决了当前的安全漏洞，使应用真正有意义。
