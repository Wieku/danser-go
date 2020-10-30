#version 330
precision highp float;

#define borderStart 0.06640625f // 34/512
#define baseBorderWidth 0.126953125f // 65/512
#define blend 0.01f

#define maxBorderWidth 1.0f - borderStart

#define slope (maxBorderWidth - baseBorderWidth) / 9

uniform vec4 col_border1;
uniform vec4 col_border;

uniform vec4 col_body1;
uniform vec4 col_body;

uniform sampler2DArray tex;

uniform float cutoff;

uniform float borderWidth;

in vec2 tex_coord;

out vec4 color;

void main()
{
    float distance = texture(tex, vec3(tex_coord, 0)).r * 2 - 1;

    if (distance >= cutoff) {
        discard;
    }

    float distance_inv = 1 - distance / cutoff;

    gl_FragDepth = 1 - distance_inv * col_border1.a;

    vec4 borderColorOuter = col_border1;
    vec4 borderColorInner = col_border;
    vec4 outerShadow = vec4(vec3(0.0f), 0.5f * distance_inv / borderStart * borderColorInner.a);

    vec4 bodyColorOuter = col_body1;
    vec4 bodyColorInner = col_body;

    float borderWidthScaled = borderWidth < 1.0f ? borderWidth * baseBorderWidth : (borderWidth - 1.0f) * slope + baseBorderWidth;
    float borderMid = borderStart + borderWidthScaled / 2;
    float borderEnd = borderStart + borderWidthScaled;

    vec4 borderColorMix = mix(borderColorOuter, borderColorInner, smoothstep(borderMid - borderWidthScaled/4, borderMid + borderWidthScaled/4, distance_inv));
    vec4 bodyColorMix = mix(bodyColorOuter, bodyColorInner, (distance_inv - borderEnd) / (1.0f - borderEnd));

    if (borderWidth < 0.01f) {
        borderColorMix = outerShadow;
    }

    if (borderWidth > 9.99f) {
        bodyColorMix = borderColorMix;
    }

    if (distance_inv <= borderStart - blend) {
        color = outerShadow;
    }

    if (distance_inv > borderStart-blend && distance_inv < borderStart+blend) {
        color = mix(outerShadow, borderColorMix, (distance_inv - (borderStart - blend)) / (2 * blend));
    }

    if (distance_inv > borderStart+blend && distance_inv <= borderEnd-blend) {
        color = borderColorMix;
    }

    if (distance_inv > borderEnd-blend && distance_inv < borderEnd+blend) {
        color = mix(borderColorMix, bodyColorMix, (distance_inv - (borderEnd - blend)) / (2 * blend));
    }

    if (distance_inv > borderEnd + blend) {
        color = bodyColorMix;
    }
}