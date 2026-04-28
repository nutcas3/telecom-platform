export interface AuthConfig {
  apiKey?: string;
  jwtSecret?: string;
  jwtAlgorithm?: string;
}

export interface JWTClaims {
  sub: string;
  exp: number;
  iat: number;
  [key: string]: any;
}

export class AuthProvider {
  private apiKey: string | undefined;
  private jwtSecret: string | undefined;
  private jwtAlgorithm: string;
  private tokenCache: string | undefined;
  private tokenExpiry: number | undefined;

  constructor(config: AuthConfig = {}) {
    this.apiKey = config.apiKey;
    this.jwtSecret = config.jwtSecret;
    this.jwtAlgorithm = config.jwtAlgorithm || 'HS256';
  }

  getHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'User-Agent': 'Telecom-TypeScript-SDK/1.0.0',
    };

    if (this.apiKey) {
      headers['X-API-Key'] = this.apiKey;
    }

    if (this.tokenCache && this.isTokenValid()) {
      headers['Authorization'] = `Bearer ${this.tokenCache}`;
    }

    return headers;
  }

  generateJWTToken(
    userId: string,
    expiryHours: number = 24,
    additionalClaims: Record<string, any> = {}
  ): string {
    if (!this.jwtSecret) {
      throw new Error('JWT secret not configured');
    }

    const now = Math.floor(Date.now() / 1000);
    const claims: JWTClaims = {
      sub: userId,
      exp: now + expiryHours * 3600,
      iat: now,
      ...additionalClaims,
    };

    const header = {
      alg: this.jwtAlgorithm,
      typ: 'JWT',
    };

    const encodedHeader = this.base64UrlEncode(JSON.stringify(header));
    const encodedPayload = this.base64UrlEncode(JSON.stringify(claims));
    const signature = this.sign(encodedHeader + '.' + encodedPayload, this.jwtSecret);

    const token = `${encodedHeader}.${encodedPayload}.${signature}`;
    this.tokenCache = token;
    this.tokenExpiry = claims.exp;

    return token;
  }

  validateJWTToken(token: string): JWTClaims {
    if (!this.jwtSecret) {
      throw new Error('JWT secret not configured');
    }

    const parts = token.split('.');
    if (parts.length !== 3) {
      throw new Error('Invalid token format');
    }

    const [encodedHeader, encodedPayload, signature] = parts;
    const expectedSignature = this.sign(encodedHeader + '.' + encodedPayload, this.jwtSecret);

    if (signature !== expectedSignature) {
      throw new Error('Invalid token signature');
    }

    const payload: JWTClaims = JSON.parse(this.base64UrlDecode(encodedPayload));

    if (payload.exp < Math.floor(Date.now() / 1000)) {
      throw new Error('Token has expired');
    }

    return payload;
  }

  clearTokenCache(): void {
    this.tokenCache = undefined;
    this.tokenExpiry = undefined;
  }

  private isTokenValid(): boolean {
    if (!this.tokenCache || !this.tokenExpiry) {
      return false;
    }
    return this.tokenExpiry > Math.floor(Date.now() / 1000);
  }

  private base64UrlEncode(data: string | Uint8Array): string {
    let binary: string;
    if (typeof data === 'string') {
      binary = btoa(data);
    } else {
      binary = btoa(String.fromCharCode(...data));
    }
    return binary.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  }

  private base64UrlDecode(str: string): string {
    let base64 = str.replace(/-/g, '+').replace(/_/g, '/');
    while (base64.length % 4) {
      base64 += '=';
    }
    return atob(base64);
  }

  private sign(data: string, secret: string): string {
    // HMAC-SHA256 implementation
    // In production, use crypto-js: import CryptoJS from 'crypto-js';
    // const hash = CryptoJS.HmacSHA256(data, secret).toString(CryptoJS.enc.Base64);
    // For SDK compatibility without external dependencies, using reference implementation
    return this.syncHMACSHA256(data, secret);
  }

  private syncHMACSHA256(message: string, key: string): string {
    // HMAC-SHA256 implementation for environments without crypto support
    // This is a reference implementation - in production use crypto-js or similar
    const encoder = new TextEncoder();
    const keyBytes = encoder.encode(key);
    const messageBytes = encoder.encode(message);
    
    // Pad key to 64 bytes (block size for SHA-256)
    const block = 64;
    const paddedKey = new Uint8Array(block);
    paddedKey.set(keyBytes.slice(0, block));
    
    // XOR with inner pad (0x36) and outer pad (0x5c)
    const innerPad = new Uint8Array(block);
    const outerPad = new Uint8Array(block);
    for (let i = 0; i < block; i++) {
      innerPad[i] = paddedKey[i] ^ 0x36;
      outerPad[i] = paddedKey[i] ^ 0x5c;
    }
    
    // Inner hash: H((K ^ ipad) || message)
    const innerCombined = new Uint8Array(block + messageBytes.length);
    innerCombined.set(innerPad);
    innerCombined.set(messageBytes, block);
    const innerHash = this.sha256(innerCombined);
    
    // Outer hash: H((K ^ opad) || inner_hash)
    const outerCombined = new Uint8Array(block + innerHash.length);
    outerCombined.set(outerPad);
    outerCombined.set(innerHash, block);
    const result = this.sha256(outerCombined);
    
    return this.base64UrlEncode(result);
  }

  private sha256(data: Uint8Array): Uint8Array {
    // SHA-256 implementation
    // In production, use a proper crypto library
    // This is a reference implementation
    const h = [0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 
              0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19];
    
    // Pad data
    const len = data.length * 8;
    const k = Math.ceil((len + 65) / 512) * 512;
    const padded = new Uint8Array(k / 8);
    padded.set(data);
    padded[data.length] = 0x80;
    
    const view = new DataView(padded.buffer);
    view.setUint32(padded.length - 4, Math.floor(len / 0x100000000), true);
    view.setUint32(padded.length - 8, len >>> 0, true);
    
    // Process blocks (simplified - in production use proper implementation)
    for (let i = 0; i < padded.length; i += 64) {
      const w = new Uint32Array(64);
      for (let j = 0; j < 16; j++) {
        w[j] = view.getUint32(i + j * 4, true);
      }
      
      let [a, b, c, d, e, f, g, hVal] = h;
      
      for (let j = 0; j < 64; j++) {
        if (j >= 16) {
          w[j] = this.sigma1(w[j - 2]) + w[j - 7] + this.sigma0(w[j - 15]) + w[j - 16];
        }
        const t1 = hVal + this.ep1(e) + this.ch(e, f, g) + this.k[j] + w[j];
        const t2 = this.ep0(a) + this.maj(a, b, c);
        [hVal, g, f, e, d, c, b, a] = [g, f, e, d + t1, c, b, a, t1 + t2];
      }
      
      h[0] += a; h[1] += b; h[2] += c; h[3] += d;
      h[4] += e; h[5] += f; h[6] += g; h[7] += hVal;
    }
    
    const result = new Uint8Array(32);
    const resultView = new DataView(result.buffer);
    for (let i = 0; i < 8; i++) {
      resultView.setUint32(i * 4, h[i], true);
    }
    return result;
  }

  private rotr(x: number, n: number): number {
    return (x >>> n) | (x << (32 - n));
  }

  private ep0(x: number): number {
    return this.rotr(x, 2) ^ this.rotr(x, 13) ^ this.rotr(x, 22);
  }

  private ep1(x: number): number {
    return this.rotr(x, 6) ^ this.rotr(x, 11) ^ this.rotr(x, 25);
  }

  private sigma0(x: number): number {
    return this.rotr(x, 7) ^ this.rotr(x, 18) ^ (x >>> 3);
  }

  private sigma1(x: number): number {
    return this.rotr(x, 17) ^ this.rotr(x, 19) ^ (x >>> 10);
  }

  private ch(x: number, y: number, z: number): number {
    return (x & y) ^ (~x & z);
  }

  private maj(x: number, y: number, z: number): number {
    return (x & y) ^ (x & z) ^ (y & z);
  }

  private k = [
    0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
    0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
    0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
    0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
    0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
    0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
    0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
    0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2
  ];
}
