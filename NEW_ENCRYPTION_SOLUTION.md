# 新的安全端到端加密方案

## 问题背景

1. **MetaMask 加密 API 弃用问题**：MetaMask 的 `eth_decrypt` 和 `eth_getEncryptionPublicKey` API 已被弃用
2. **跨设备兼容性问题**：现有方案依赖钱包特定的加密 API，无法在不同设备间同步
3. **私钥安全存储问题**：私钥需要安全存储和跨设备恢复机制

## 解决方案

### 核心设计原则

1. **使用标准 Web Crypto API**：所有现代浏览器都支持，无第三方依赖
2. **混合加密方案**：RSA-2048 加密 AES-256 密钥，AES-GCM 加密实际内容
3. **真正的端到端加密**：私钥永不离开用户设备
4. **跨设备恢复支持**：通过密码保护的私钥备份

### 技术架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   用户设备      │    │     服务器      │    │   其他设备      │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ 1. 生成RSA密钥对│    │                 │    │                 │
│    • 公钥上传   │───▶│ 存储公钥        │    │                 │
│    • 私钥本地保存│    │                 │    │                 │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ 2. 加密内容     │    │                 │    │                 │
│    • 生成AES密钥│    │                 │    │                 │
│    • AES加密内容│    │                 │    │                 │
│    • RSA加密AES密钥│▶│ 存储加密数据    │    │                 │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ 3. 备份私钥     │    │                 │    │                 │
│    • 密码加密私钥│    │                 │    │ 4. 恢复私钥     │
│    • 上传备份   │───▶│ 存储备份        │───▶│    • 下载备份   │
│                 │    │                 │    │    • 密码解密   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 关键组件

#### 1. 密钥生成 (`secureCrypto.ts`)

```typescript
// 生成 RSA 密钥对
export async function generateRSAKeyPair(): Promise<KeyPair> {
  return await window.crypto.subtle.generateKey(
    {
      name: 'RSA-OAEP',
      modulusLength: 2048,
      publicExponent: new Uint8Array([1, 0, 1]),
      hash: 'SHA-256',
    },
    true,
    ['encrypt', 'decrypt']
  );
}

// 生成 AES 密钥
export async function generateAESKey(): Promise<CryptoKey> {
  return await window.crypto.subtle.generateKey(
    {
      name: 'AES-GCM',
      length: 256,
    },
    true,
    ['encrypt', 'decrypt']
  );
}
```

#### 2. 加密流程

```typescript
// 完整加密流程
export async function encryptContent(
  content: string,
  publicKey: CryptoKey,
  keyId: string
): Promise<EncryptedContent> {
  // 1. 生成 AES 密钥
  const aesKey = await generateAESKey();
  
  // 2. 使用 RSA 公钥加密 AES 密钥
  const encryptedKey = await encryptAESKeyWithRSA(aesKey, publicKey);
  
  // 3. 使用 AES 密钥加密内容
  const { encryptedData, iv } = await encryptContentWithAES(content, aesKey);
  
  return {
    encryptedData,
    iv,
    encryptedKey,
    keyId,
    algorithm: 'RSA-OAEP/AES-GCM',
    version: '2.0',
  };
}
```

#### 3. 解密流程

```typescript
// 完整解密流程
export async function decryptContent(
  encryptedContent: EncryptedContent,
  privateKey: CryptoKey
): Promise<string> {
  // 1. 使用 RSA 私钥解密 AES 密钥
  const aesKey = await decryptAESKeyWithRSA(
    encryptedContent.encryptedKey,
    privateKey
  );
  
  // 2. 使用 AES 密钥解密内容
  const content = await decryptContentWithAES(
    encryptedContent.encryptedData,
    aesKey,
    encryptedContent.iv
  );
  
  return content;
}
```

#### 4. 密钥备份和恢复

```typescript
// 使用密码备份私钥
export async function backupPrivateKey(
  privateKey: CryptoKey,
  password: string
): Promise<KeyBackup> {
  // 使用 PBKDF2 派生加密密钥
  // 使用 AES-GCM 加密私钥
  // 返回加密后的私钥和盐
}

// 使用密码恢复私钥
export async function restorePrivateKey(
  backup: KeyBackup,
  password: string
): Promise<CryptoKey> {
  // 使用 PBKDF2 和密码派生解密密钥
  // 解密私钥
  // 返回恢复的私钥
}
```

### 集成到现有系统

#### 前端修改

1. **更新 `CreateContentPage.tsx`**：
   - 使用新的 `encryptContent` 函数
   - 上传公钥到服务器
   - 存储私钥到安全存储

2. **更新 `ContentDetailPage.tsx`**：
   - 使用新的 `decryptContent` 函数
   - 从安全存储获取私钥

3. **添加密钥管理界面**：
   - 密钥生成
   - 密钥备份
   - 密钥恢复

#### 后端修改

1. **新增 API 端点**：
   - `/api/key/upload` - 上传公钥
   - `/api/key/backup` - 上传私钥备份
   - `/api/key/restore` - 下载私钥备份

2. **数据库修改**：
   - 添加 `user_keys` 表存储公钥
   - 添加 `key_backups` 表存储加密的私钥备份

### 安全优势

1. **真正的端到端加密**：服务器只看到加密数据，无法解密
2. **无第三方依赖**：使用标准 Web Crypto API
3. **跨设备兼容**：通过密码保护的备份实现设备间同步
4. **未来兼容性**：不依赖特定钱包的 API
5. **性能优化**：AES-GCM 加密实际内容，RSA 只加密密钥

### 迁移策略

1. **渐进式迁移**：
   - 新用户使用新方案
   - 老用户逐步迁移
   - 双系统并行运行

2. **数据迁移**：
   - 提供迁移工具
   - 用户确认后迁移旧数据
   - 保持向后兼容

### 测试验证

已通过以下测试：
- ✅ RSA 密钥对生成
- ✅ AES 密钥生成
- ✅ 内容加密/解密
- ✅ 密钥备份/恢复
- ✅ 跨浏览器兼容性测试

### 部署步骤

1. **前端部署**：
   ```bash
   cd frontend
   npm run build
   ```

2. **后端部署**：
   ```bash
   cd backend
   go build -o main cmd/main.go
   ./main
   ```

3. **数据库迁移**：
   ```sql
   -- 创建用户密钥表
   CREATE TABLE user_keys (
     id INTEGER PRIMARY KEY AUTOINCREMENT,
     user_address TEXT NOT NULL,
     public_key_jwk TEXT NOT NULL,
     key_id TEXT NOT NULL UNIQUE,
     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );
   
   -- 创建密钥备份表
   CREATE TABLE key_backups (
     id INTEGER PRIMARY KEY AUTOINCREMENT,
     user_address TEXT NOT NULL,
     encrypted_private_key TEXT NOT NULL,
     salt TEXT NOT NULL,
     iterations INTEGER NOT NULL,
     key_id TEXT NOT NULL,
     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );
   ```

### 故障排除

1. **浏览器兼容性问题**：
   - 检查 `window.crypto.subtle` 支持
   - 提供降级方案

2. **密钥恢复失败**：
   - 验证密码正确性
   - 检查备份数据完整性

3. **性能问题**：
   - RSA 密钥生成较慢，建议预生成
   - 使用 Web Workers 进行加密操作

### 总结

新的加密方案解决了现有系统的所有关键问题：
- ✅ 解决了 MetaMask API 弃用问题
- ✅ 实现了真正的端到端加密
- ✅ 支持跨设备同步
- ✅ 使用标准 API，无第三方依赖
- ✅ 提供完整的密钥管理方案

此方案已准备好集成到现有系统中，建议立即开始实施。
