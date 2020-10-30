#version 330

#define INVSQ2PI 0.398942

uniform sampler2DArray tex;

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
    float tSigma = length(direction*sigma);

    float gs = gauss(0.0, tSigma);

    vec4 inc = texture(tex, vec3(tex_coord, 0));

    color = inc*gs;

    float totalGauss = gs;

    int kSize = int(length(kernelSize*direction));

    for (int i = 2; i < 200; i+=2) {
        float fac = float(i) - 0.5;

        gs = gauss(i, tSigma)*2.0;
        totalGauss += 2.0*gs;

        vec2 mv = fac * direction / size;

        color += texture(tex, vec3(tex_coord + mv, 0)) * gs;
        color += texture(tex, vec3(tex_coord - mv, 0)) * gs;

        if (i >= kSize) {
            break;
        }

    }

    color /= totalGauss;
}