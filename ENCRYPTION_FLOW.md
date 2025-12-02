# VaultSeed 加解密流程详解

本文档详细说明 VaultSeed 前端加解密的完整流程。

## 当前实现状态

**重要说明**：当前实现中，对称密钥的加密/解密部分使用了简化处理（base64 编码），而不是真正的非对称加密。这是为了在浏览器环境中快速实现原型。在生产环境中，应该使用真正的非对称加密算法（如 RSA-OAEP 或 ECIES）。

## 加密流程（创建内容）

### 1. 用户输入
- 标题（明文，存储在后端）
- 内容（≤256 字符，前端加密）

### 2. 生成对称密钥
```typescript
// 生成 AES-256 密钥
const aesKey = await generateAESKey();
const aesKeyString = await exportAESKey(aesKey);
```

**实现细节**：
- 优先使用 Web Crypto API（如果可用）
- 备选方案：crypto-js 库
- 密钥长度：256 bits
- 算法：AES-GCM（Web Crypto）或 AES-CBC（crypto-js）

### 3. 加密内容
```typescript
const { encryptedData, iv } = await encryptWithAES(formData.content, aesKey);
```

**实现细节**：
- **Web Crypto API**：AES-GCM 模式，12字节随机 IV
- **crypto-js**：AES-CBC 模式，PKCS7 填充，12字节随机 IV
- 输出：十六进制字符串格式的加密数据和 IV

### 4. 加密对称密钥（当前简化实现）
```typescript
const encryptedKey = await encryptKeyWithPublicKey(aesKeyString, publicKey);
```

**当前实现问题**：
```typescript
// 简化处理 - 不是真正的非对称加密！
export async function encryptKeyWithPublicKey(
  aesKey: string,
  publicKey: string
): Promise<string> {
  // 注意：这里简化处理，实际应该使用非对称加密
  // 由于浏览器限制，这里只做 base64 编码
  return btoa(aesKey + ':' + publicKey.substring(0, 32));
}
```

**问题**：这只是一个 base64 编码，不是真正的加密。任何人都可以解码并获取 AES 密钥。

### 5. 发送到后端
```typescript
const response = await contentAPI.create(
  formData.title,          // 明文标题
  encryptedKey,            // "加密"后的对称密钥（实际是 base64 编码）
  iv,                      // 初始化向量
  encryptedData            // 加密后的内容
);
```

### 6. 本地存储（可选）
```typescript
const localContents = JSON.parse(localStorage.getItem('localContents') || '[]');
localContents.push({
  id: response.data.id,
  title: formData.title,
  encryptedData,
  iv,
  createdAt: new Date().toISOString(),
});
localStorage.setItem('localContents', JSON.stringify(localContents));
```

## 解密流程（查看内容）

### 1. 获取内容详情
```typescript
const response = await contentAPI.getDetail(parseInt(id!));
// 返回包含 nonce 的内容信息
```

### 2. 钱包签名
```typescript
const message = generateDecryptMessage(parseInt(id!), nonce);
const signature = await signer.signMessage(message);
```

### 3. 请求解密
```typescript
const response = await contentAPI.decrypt(parseInt(id!), signature, message, nonce);
// 后端验证签名后返回加密数据
```

### 4. 解密对称密钥（当前简化实现）
```typescript
const aesKeyString = await decryptKeyWithPrivateKey(encrypted_key, privateKey);
```

**当前实现问题**：
```typescript
export async function decryptKeyWithPrivateKey(
  encryptedKey: string,
  privateKey: string
): Promise<string> {
  // 简化处理
  const decoded = atob(encryptedKey);
  return decoded.split(':')[0];
}
```

**问题**：这只是 base64 解码，不是真正的解密。

### 5. 解密内容
```typescript
const aesKey = await importAESKey(aesKeyString);
const decrypted = await decryptWithAES(encrypted_data, aesKey, iv);
```

## 安全问题和改进建议

### 当前问题

1. **对称密钥加密不安全**：
   - 使用 base64 编码而不是真正的非对称加密
   - 任何人都可以解码获取 AES 密钥
   - 违背了端到端加密的原则

2. **密钥管理**：
   - 私钥存储在 localStorage 中
   - 没有硬件钱包集成

3. **算法一致性**：
   - Web Crypto API 和 crypto-js 使用不同模式（GCM vs CBC）

### 改进方案

#### 方案 1：使用 Web Crypto API 的非对称加密
```typescript
// 生成 RSA-OAEP 密钥对
const keyPair = await window.crypto.subtle.generateKey(
  {
    name: "RSA-OAEP",
    modulusLength: 2048,
    publicExponent: new Uint8Array([1, 0, 1]),
    hash: "SHA-256",
  },
  true,
  ["encrypt", "decrypt"]
);

// 使用公钥加密 AES 密钥
const encryptedAesKey = await window.crypto.subtle.encrypt(
  {
    name: "RSA-OAEP",
  },
  keyPair.publicKey,
  aesKeyBytes
);

// 使用私钥解密 AES 密钥
const decryptedAesKey = await window.crypto.subtle.decrypt(
  {
    name: "RSA-OAEP",
  },
  keyPair.privateKey,
  encryptedAesKey
);
```

#### 方案 2：使用以太坊钱包的加密
```typescript
// 使用 eth-sig-util 或类似库
import { encrypt, decrypt } from 'eth-sig-util';

// 使用钱包公钥加密
const encrypted = encrypt(
  publicKey,
  { data: aesKeyString },
  'x25519-xsalsa20-poly1305'
);

// 使用钱包私钥解密
const decrypted = decrypt(encrypted, privateKey);
```

#### 方案 3：使用 libsodium-wrappers（推荐）
```typescript
import sodium from 'libsodium-wrappers';

await sodium.ready;

// 使用 X25519 密钥交换
const keyPair = sodium.crypto_box_keypair();
const encrypted = sodium.crypto_box_seal(aesKeyString, keyPair.publicKey);
const decrypted = sodium.crypto_box_seal_open(encrypted, keyPair.publicKey, keyPair.privateKey);
```

## 立即修复建议

### 短期修复（最小改动）
1. 至少使用一个简单的加密算法，而不是 base64
2. 使用钱包地址作为密钥进行对称加密

```typescript
// 改进的简单加密
export async function encryptKeyWithPublicKey(
  aesKey: string,
  publicKey: string
): Promise<string> {
  // 使用公钥的哈希作为密钥进行 AES 加密
  const key = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(publicKey));
  const cryptoKey = await crypto.subtle.importKey('raw', key, 'AES-GCM', false, ['encrypt']);
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const encrypted = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv },
    cryptoKey,
    new TextEncoder().encode(aesKey)
  );
  return JSON.stringify({
    iv: Array.from(iv),
    data: Array.from(new Uint8Array(encrypted))
  });
}
```

### 长期改进
1. 集成硬件钱包支持
2. 实现真正的端到端加密
3. 添加密钥轮换机制
4. 实现前向保密

## 测试验证步骤

### 验证当前实现
```bash
# 1. 创建加密内容
# 2. 检查网络请求，查看发送到后端的数据
# 3. 尝试解码 base64 加密的密钥
# 4. 验证是否能直接解密内容

# 示例解码
echo "base64_encoded_string" | base64 -d
```

### 验证改进后的实现
1. 创建内容并加密
2. 检查网络请求，确保密钥是真正加密的
3. 尝试直接解密失败（证明加密有效）
4. 正常流程解密成功

## 总结

当前实现的主要问题是**对称密钥的加密使用了 base64 编码而不是真正的非对称加密**。这使得系统实际上不是真正的端到端加密，因为服务器可以解码并获取 AES 密钥。

**建议立即修复**：至少使用一个简单的对称加密算法来保护 AES 密钥，而不是使用 base64 编码。

**长期目标**：实现真正的非对称加密，使用 Web Crypto API 的 RSA-OAEP 或集成 libsodium 进行更安全的密钥管理。

## 已修复的问题："meta收到的不是密文而是JSON"

### 问题描述
在使用钱包加密方案时，MetaMask 的 `eth_decrypt` API 期望接收密文字符串，但实际收到的是 JSON 格式的数据，导致解密失败。

### 问题根源
1. `encryptWithWalletPublicKey` 函数返回的是对象，而不是字符串
2. `decryptWithWallet` 函数没有正确处理各种输入格式
3. 传递给 MetaMask `eth_decrypt` 的是 JSON 对象而不是密文字符串

### 修复方案
已对 `frontend/src/utils/crypto.ts` 文件进行以下修复：

#### 1. 修复 `encryptWithWalletPublicKey` 函数
```typescript
// 修复前：返回对象
return encryptedData; // 对象

// 修复后：返回 JSON 字符串
return JSON.stringify(encryptedData); // 字符串
```

#### 2. 修复 `decryptWithWallet` 函数
- 添加了对各种输入格式的处理逻辑
- 确保传递给 MetaMask `eth_decrypt` 的是正确的密文字符串
- 支持版本化的加密数据格式

#### 3. 修复 `encryptAESKeyWithWallet` 函数
- 确保返回版本化的 JSON 数据
- `encryptedKey` 字段已经是 JSON 字符串

#### 4. 修复 `decryptAESKeyWithWallet` 函数
- 根据版本选择解密方法
- 正确处理版本 2.0（钱包加密）和版本 1.1（本地加密）

### 修复后的数据格式
#### 版本 2.0（钱包加密）
```json
{
  "version": "2.0",
  "algorithm": "x25519-xsalsa20-poly1305",
  "encryptedKey": "{\"version\":\"x25519-xsalsa20-poly1305\",\"nonce\":\"...\",\"ephemPublicKey\":\"...\",\"ciphertext\":\"...\"}",
  "timestamp": "2025-12-02T03:02:03.187Z"
}
```

#### 版本 1.1（本地加密）
```json
{
  "version": "1.1",
  "algorithm": "AES-GCM-SHA256",
  "encryptedKey": "{\"iv\":[...],\"data\":[...],\"algorithm\":\"AES-GCM-SHA256\"}",
  "timestamp": "2025-12-02T03:02:03.187Z"
}
```

### 测试验证
已通过测试验证修复有效：
1. 加密流程生成正确的 JSON 格式数据
2. 解密流程能正确处理各种输入格式
3. MetaMask 能正确解密 AES 密钥
4. 兼容旧格式数据（base64 编码）

### 总结
此修复解决了"meta收到的不是密文而是JSON"的问题，确保了钱包加密/解密流程的正常工作。现在系统能正确处理：
- ✅ 钱包公钥加密 AES 密钥
- ✅ MetaMask 解密 AES 密钥
- ✅ 版本化的加密数据格式
- ✅ 向后兼容性
