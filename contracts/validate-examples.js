#!/usr/bin/env node

const fs = require('fs/promises');
const { existsSync } = require('fs');
const path = require('path');

const contractsNodeModules = path.join(__dirname, 'node_modules');
const ajvPath = path.join(contractsNodeModules, 'ajv');
const ajvFormatsPath = path.join(contractsNodeModules, 'ajv-formats');

if (!existsSync(ajvPath) || !existsSync(ajvFormatsPath)) {
  console.error(
    'Missing dependencies. Please install contracts tooling with `npm --prefix contracts install --no-fund --no-audit` before running validation.',
  );
  process.exit(1);
}

const Ajv = require('ajv');
const addFormats = require('ajv-formats');

const EXAMPLES_ROOT = path.join(__dirname, 'examples');
const SCHEMAS_ROOT = path.join(__dirname, 'schemas');

const ajv = new Ajv({ allErrors: true, strict: false });
addFormats(ajv);

async function fileExists(targetPath) {
  try {
    await fs.access(targetPath);
    return true;
  } catch {
    return false;
  }
}

async function listExampleFiles(dir) {
  const exists = await fileExists(dir);
  if (!exists) {
    return [];
  }

  const entries = await fs.readdir(dir, { withFileTypes: true });
  const files = [];

  for (const entry of entries) {
    const entryPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      files.push(...(await listExampleFiles(entryPath)));
    } else if (entry.isFile() && entry.name.endsWith('.json')) {
      files.push(entryPath);
    }
  }

  return files.sort();
}

function normalizeSchemaRef(schemaRef) {
  let ref = schemaRef.trim();
  try {
    const url = new URL(ref);
    ref = url.pathname.slice(1);
  } catch {
    // Not a URL; ignore.
  }

  return ref.replace(/^\.\//, '').replace(/^\//, '');
}

async function resolveSchemaPath(schemaRef) {
  const normalized = normalizeSchemaRef(schemaRef);
  const candidates = [];

  if (path.isAbsolute(schemaRef)) {
    candidates.push(schemaRef);
  }

  if (schemaRef.endsWith('.json')) {
    candidates.push(path.join(SCHEMAS_ROOT, normalized));
  } else {
    candidates.push(path.join(SCHEMAS_ROOT, `${normalized}.json`));
    candidates.push(path.join(SCHEMAS_ROOT, normalized, 'schema.json'));
  }

  const uniqueCandidates = Array.from(new Set(candidates));

  for (const candidate of uniqueCandidates) {
    if (await fileExists(candidate)) {
      return candidate;
    }
  }

  const locations = uniqueCandidates
    .map((candidate) => ` - ${path.relative(process.cwd(), candidate)}`)
    .join('\n');

  throw new Error(
    `Could not locate schema for identifier "${schemaRef}". Looked in:\n${locations}`,
  );
}

const validatorCache = new Map();

async function loadValidator(schemaPath) {
  if (validatorCache.has(schemaPath)) {
    return validatorCache.get(schemaPath);
  }

  let schemaJson;
  try {
    const rawSchema = await fs.readFile(schemaPath, 'utf8');
    schemaJson = JSON.parse(rawSchema);
  } catch (error) {
    throw new Error(`Unable to read schema at ${schemaPath}: ${error.message}`);
  }

  let validator;
  try {
    validator = ajv.compile(schemaJson);
  } catch (error) {
    throw new Error(`Failed to compile schema ${schemaPath}: ${error.message}`);
  }

  validatorCache.set(schemaPath, validator);
  return validator;
}

function formatErrors(errors = []) {
  return errors
    .map((err) => {
      const location = err.instancePath && err.instancePath.length > 0 ? err.instancePath : '/';
      const params = err.params && Object.keys(err.params).length > 0 ? ` (${JSON.stringify(err.params)})` : '';
      return ` - ${location} ${err.message}${params}`;
    })
    .join('\n');
}

function indent(text, prefix = '  ') {
  return text
    .split('\n')
    .map((line) => `${prefix}${line}`)
    .join('\n');
}

async function validateExamples() {
  const exampleFiles = await listExampleFiles(EXAMPLES_ROOT);

  if (exampleFiles.length === 0) {
    console.log('No contract example JSON files found.');
    return;
  }

  const failures = [];

  for (const filePath of exampleFiles) {
    try {
      const raw = await fs.readFile(filePath, 'utf8');
      let envelope;
      try {
        envelope = JSON.parse(raw);
      } catch (error) {
        throw new Error(`Invalid JSON: ${error.message}`);
      }

      if (!envelope || typeof envelope !== 'object' || Array.isArray(envelope)) {
        throw new Error('Example must be a JSON object.');
      }

      const schemaRef = envelope.schema;
      if (typeof schemaRef !== 'string' || schemaRef.trim().length === 0) {
        throw new Error('Missing schema identifier in "schema" field.');
      }

      const schemaPath = await resolveSchemaPath(schemaRef);
      const validator = await loadValidator(schemaPath);
      const valid = validator(envelope);

      if (!valid) {
        const formattedErrors = formatErrors(validator.errors);
        throw new Error(
          `Schema validation failed using ${path.relative(process.cwd(), schemaPath)}:\n${formattedErrors}`,
        );
      }

      console.log(`✔ ${path.relative(process.cwd(), filePath)} validated`);
    } catch (error) {
      failures.push({ filePath, message: error.message });
    }
  }

  if (failures.length > 0) {
    console.error(`\nValidation failed for ${failures.length} example(s):`);
    for (const { filePath, message } of failures) {
      console.error(`- ${path.relative(process.cwd(), filePath)}`);
      console.error(indent(message, '    '));
    }
    process.exit(1);
  }

  console.log(`\nValidated ${exampleFiles.length} example(s).`);
}

validateExamples().catch((error) => {
  console.error(error);
  process.exit(1);
});
