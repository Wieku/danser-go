#version 330

#define INVSQ2PI 0.398942

uniform sampler2D tex;

uniform vec2 kernelSize;
uniform vec2 direction;
uniform vec2 sigma;
uniform vec2 size;

in vec2 tex_coord;

out vec4 color;

float gauss(float x, float sigma) {
    return INVSQ2PI * exp(-0.5 * x * x / (sigma * sigma)) / sigma;
}

void main() {
    color = vec4(0.0);

    float tSigma = length(direction*sigma);
    //float size = length(direction*size);

    float totalGauss = 0.0f;

    float gs = gauss(0, tSigma);

    color += texture2D(tex, tex_coord)*gs;

    totalGauss += gs;

    int kSize = int(length(kernelSize*direction));

    for (int i = 2; i < kSize; i+=2) {
        float fac = float(i) - 0.5f;

        gs = gauss(i, tSigma)*2;
        totalGauss += 2*gs;

        vec2 mv = fac * direction / size;

        color += texture2D(tex, tex_coord + mv) * gs;
        color += texture2D(tex, tex_coord - mv) * gs;
    }

    color /= totalGauss;
}